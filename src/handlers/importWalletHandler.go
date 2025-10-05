package handlers

import (
	"CypherD-hackathon/src/internal/models"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ImportWalletHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Decode the incoming JSON request
		var req CreateWalletRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Wallet address is required"})
			return
		}

		rand.Seed(time.Now().UnixNano())
		balance := 1.0 + rand.Float64()*(10.0-1.0)
		walletBalance := fmt.Sprintf("%.18f", balance)

		// 2. Find the wallet or create it with a wallet balance
		var wallet models.Wallet

		result := db.Where(models.Wallet{Address: req.Address}).
			Attrs(models.Wallet{Balance: walletBalance, Email: req.Email}).
			FirstOrCreate(&wallet)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// 3. If a new wallet was created, assign a random balance
		if result.RowsAffected > 0 {
			// Respond with 201 Created
			c.JSON(http.StatusCreated, wallet)
			return
		}

		// If the wallet already existed, just return its data
		c.JSON(http.StatusOK, wallet)
	}
}
