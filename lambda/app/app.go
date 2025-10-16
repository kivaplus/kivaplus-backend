package app

import (
	"log"

	"github.com/kivaplus/kivaplus-backend/lambda/api"
	"github.com/kivaplus/kivaplus-backend/lambda/database"
)

type App struct {
	ApiHandler api.ApiHandler
}

func NewApp() App {
	//init dbStore
	db, err := database.NewUserStore()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	apiHandler := api.NewApiHandler(db)

	return App{
		ApiHandler: apiHandler,
	}
}
