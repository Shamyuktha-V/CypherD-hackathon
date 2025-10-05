package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync"
	"time"

	"CypherD-hackathon/src/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var pendingTransactions = make(map[string]PendingTransaction)
var pendingTxMutex = &sync.Mutex{}

func InitiateTransferHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req InitiateTransferRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		var amountETH *big.Int
		var amountUSD *string
		var message string

		if req.Currency == "USD" {
			usdAmount, _, _ := new(big.Float).Parse(req.Amount, 10)
			usdAmountSmallestUnit := new(big.Int)
			new(big.Float).Mul(usdAmount, big.NewFloat(1e6)).Int(usdAmountSmallestUnit)

			ethFromAPI, err := getEthQuoteFromSkipAPI(usdAmountSmallestUnit.String())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get price quote"})
				return
			}
			amountETH = ethFromAPI
			amountUSD = &req.Amount

			ethFloat := new(big.Float).SetInt(amountETH)
			ethFloat.Quo(ethFloat, big.NewFloat(1e18))
			message = fmt.Sprintf("Transfer %s ETH ($%s USD) to %s", ethFloat.Text('f', 4), req.Amount, req.To)
		} else {
			ethAmount, _, _ := new(big.Float).Parse(req.Amount, 10)
			amountETH = new(big.Int)
			new(big.Float).Mul(ethAmount, big.NewFloat(1e18)).Int(amountETH)
			message = fmt.Sprintf("Transfer %s ETH to %s", req.Amount, req.To)
		}

		hash := "0x" + generateTransactionHash(req.From, req.To)

		newTransaction := models.Transaction{
			SenderAddress:    req.From,
			RecipientAddress: req.To,
			AmountETH:        amountETH.String(),
			AmountUSD:        amountUSD,
			Status:           "pending",
			TransactionHash:  hash,
		}

		if result := db.Create(&newTransaction); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create transaction record"})
			return
		}

		c.JSON(http.StatusOK, InitiateTransferResponse{
			MessageToSign: message,
			TransactionID: newTransaction.ID,
		})
	}
}

func generateTransactionHash(sender string, recipient string) string {
	data := fmt.Sprintf("%s-%s-%d", sender, recipient, time.Now().UnixNano())
	hash := crypto.Keccak256Hash([]byte(data))
	return hex.EncodeToString(hash.Bytes())
}

func ExecuteTransferHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExecuteTransferRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		var pendingTx models.Transaction
		if err := db.First(&pendingTx, "id = ? AND status = ?", req.TransactionID, "pending").Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending transaction not found or has expired"})
			return
		}

		var message string

		amountInt, ok := new(big.Int).SetString(pendingTx.AmountETH, 10)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not parse transaction amount"})
			return
		}

		ethFloat := new(big.Float).SetInt(amountInt)
		ethFloat.Quo(ethFloat, big.NewFloat(1e18))

		if pendingTx.AmountUSD != nil {
			message = fmt.Sprintf("Transfer %s ETH ($%s USD) to %s", ethFloat.Text('f', 4), *pendingTx.AmountUSD, pendingTx.RecipientAddress)
		} else {
			originalAmountString := ethFloat.Text('g', -1)
			message = fmt.Sprintf("Transfer %s ETH to %s", originalAmountString, pendingTx.RecipientAddress)
		}

		isValid, err := verifySignature(pendingTx.SenderAddress, req.Signature, []byte(message))
		if err != nil || !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature", "expectedMessage": message})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			var senderWallet, recipientWallet models.Wallet
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&senderWallet, "address = ?", pendingTx.SenderAddress).Error; err != nil {
				return fmt.Errorf("sender not found")
			}
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&recipientWallet, "address = ?", pendingTx.RecipientAddress).Error; err != nil {
				return fmt.Errorf("recipient not found")
			}

			transferAmount, ok := new(big.Int).SetString(pendingTx.AmountETH, 10)
			if !ok {
				return fmt.Errorf("could not parse transfer amount")
			}

			senderBalanceFloat, _, err := new(big.Float).Parse(senderWallet.Balance, 10)
			if err != nil {
				return fmt.Errorf("could not parse sender balance string")
			}

			weiConversion := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
			senderBalanceWei := new(big.Int)
			new(big.Float).Mul(senderBalanceFloat, weiConversion).Int(senderBalanceWei)

			if senderBalanceWei.Cmp(transferAmount) < 0 {
				return fmt.Errorf("insufficient funds")
			}

			senderBalanceWei.Sub(senderBalanceWei, transferAmount)

			recipientBalanceFloat, _, err := new(big.Float).Parse(recipientWallet.Balance, 10)
			if err != nil {
				return fmt.Errorf("could not parse recipient balance string")
			}
			recipientBalanceWei := new(big.Int)
			new(big.Float).Mul(recipientBalanceFloat, weiConversion).Int(recipientBalanceWei)
			recipientBalanceWei.Add(recipientBalanceWei, transferAmount)

			finalSenderBalanceETH := new(big.Float).SetInt(senderBalanceWei)
			finalSenderBalanceETH.Quo(finalSenderBalanceETH, weiConversion)

			finalRecipientBalanceETH := new(big.Float).SetInt(recipientBalanceWei)
			finalRecipientBalanceETH.Quo(finalRecipientBalanceETH, weiConversion)

			senderWallet.Balance = finalSenderBalanceETH.Text('f', 18)
			recipientWallet.Balance = finalRecipientBalanceETH.Text('f', 18)

			if err := tx.Save(&senderWallet).Error; err != nil {
				return err
			}
			if err := tx.Save(&recipientWallet).Error; err != nil {
				return err
			}

			pendingTx.Status = "completed"
			if err := tx.Save(&pendingTx).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Transaction completed"})
	}
}

func getEthQuoteFromSkipAPI(usdAmountInSmallestUnit string) (*big.Int, error) {
	apiURL := "https://api.skip.build/v2/fungible/msgs_direct"
	payload := map[string]interface{}{
		"source_asset_denom":         "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		"source_asset_chain_id":      "1",
		"dest_asset_denom":           "ethereum-native",
		"dest_asset_chain_id":        "1",
		"amount_in":                  usdAmountInSmallestUnit,
		"chain_ids_to_addresses":     map[string]string{"1": "0x742d35Cc6634C0532925a3b8D4C9db96c728b0B4"},
		"slippage_tolerance_percent": "1",
		"smart_swap_options":         map[string]bool{"evm_swaps": true},
		"allow_unsafe":               false,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Skip API returned non-200 status: %d - %s", resp.StatusCode, string(respBody))
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var apiResp SkipAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, err
	}

	amountOutString := apiResp.Route.AmountOut
	if amountOutString == "" {
		return nil, fmt.Errorf("Skip API did not return an amount_out value in the route object")
	}

	amountOut, success := new(big.Int).SetString(amountOutString, 10)
	if !success {
		return nil, fmt.Errorf("could not parse amount_out from Skip API: %s", amountOutString)
	}

	return amountOut, nil
}

func verifySignature(fromAddress, sigHex string, msg []byte) (bool, error) {
	sig, err := hexutil.Decode(sigHex)
	if err != nil {
		return false, err
	}

	if sig[64] != 27 && sig[64] != 28 {
		sig[64] += 27
	}

	msgHash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)),
	)

	pubKey, err := crypto.SigToPub(msgHash.Bytes(), sig)
	if err != nil {
		return false, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	return fromAddress == recoveredAddr.Hex(), nil
}

func ViewTransactionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Address parameter is required"})
			return
		}
		var transactions []models.Transaction

		result := db.Where("sender_address = ? OR recipient_address = ?", address, address).
			Order("created_at desc").
			Find(&transactions)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}
