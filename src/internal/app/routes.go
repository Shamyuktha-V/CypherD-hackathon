package app

import (
	"CypherD-hackathon/src/handlers"
	appConfig "CypherD-hackathon/src/internal/config"
)

func mapURLs(appConfig *appConfig.AppConfig) {
	db := appConfig.DB

	// Register the handler for the specific routes

	// create wallet handler
	createWalletHandler := handlers.CreateWalletHandler(db)
	router.POST(CreatewalletURL, createWalletHandler)

	// get wallet handler
	getWalletHandler := handlers.GetWalletHandler(db)
	router.GET(GetWalletURL, getWalletHandler)

	// imort wallet handler
	importWallerHandler := handlers.ImportWalletHandler(db)
	router.POST(ImportWalletURL, importWallerHandler)

}
