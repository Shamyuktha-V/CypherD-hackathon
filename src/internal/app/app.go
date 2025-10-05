package app

import (
	configuration "CypherD-hackathon/src/internal/config"
	"CypherD-hackathon/src/internal/models"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"log"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var router = gin.New()

func StartApplication(appConf *configuration.AppConfig) {
	router.Use(gin.Recovery())

	// Initialize database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", appConf.DBConfig.DBUSER, appConf.DBConfig.DBPASSWORD, appConf.DBConfig.DBHOST, appConf.DBConfig.DBPORT, appConf.DBConfig.DBNAME)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	appConf.DB = db
	err = db.AutoMigrate(models.Wallet{}, models.Transaction{})
	if err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}

	// REST endpoint

	mapURLs(appConf)

	// Start the Gin server
	log.Printf("REST server is running on port 8080")

	// Start the Gin server in a separate goroutine
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	// The main goroutine will terminate after executing all lines of code.
	// To keep the session alive, we use signal handling to listen for interrupt or termination signals.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-sig
	fmt.Println("Shutting down server...")

}
