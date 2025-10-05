package handlers

import (
	"CypherD-hackathon/src/internal/models"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateWalletHandler creates a new wallet entry in the database using Gin.
func CreateWalletHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Decode the incoming JSON request using Gin's binding
		var req CreateWalletRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Wallet address is required"})
			return
		}

		// 2. Create a new wallet model with a random balance
		rand.Seed(time.Now().UnixNano())
		balance := 1.0 + rand.Float64()*(10.0-1.0)

		newWallet := models.Wallet{
			Address: req.Address,
			Email:   req.Email,
			Balance: fmt.Sprintf("%.18f", balance),
		}

		// 3. Save the new wallet to the database
		result := db.Create(&newWallet)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "Duplicate entry") {
				c.JSON(http.StatusConflict, gin.H{"error": "Wallet address already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		// 4. Send a success response using Gin's JSON method
		log.Printf("Successfully created wallet for address: %s", newWallet.Address)
		c.JSON(http.StatusCreated, newWallet)
	}
}
