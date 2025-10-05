package handlers

import (
	"CypherD-hackathon/src/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func GetWalletHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the address from the URL parameter
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Address parameter is required"})
			return
		}

		// 2. Query the database for the wallet
		var wallet models.Wallet
		result := db.Where("address = ?", address).First(&wallet)

		// 3. Handle errors, specifically "record not found"
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
				return
			}
			// For any other database error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// 4. Send a success response with the wallet data
		c.JSON(http.StatusOK, wallet)
	}
}
