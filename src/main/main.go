package main

import (
	"fmt"

	"log"

	"CypherD-hackathon/src/internal/app"
	configuration "CypherD-hackathon/src/internal/config"

	"github.com/joho/godotenv"
)

var appConfig *configuration.AppConfig

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Println("error loading env file"+" : %s", err)
	}

	appConfig = configuration.NewConfig()
}

func main() {
	fmt.Println("Connecting to application")
	app.StartApplication(appConfig)
}
