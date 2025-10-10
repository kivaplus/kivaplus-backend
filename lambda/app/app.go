package app

import (
	"github.com/kivaplus/kivaplus-backend/lambda/api"
	"github.com/kivaplus/kivaplus-backend/lambda/database"
)

type App struct {
	ApiHandler api.ApiHandler
}

func NewApp() App {
	//init dbStore
	db := database.NewDynamoDBClient()
	apiHandler := api.NewApiHandler(db)

	return App{
		ApiHandler: apiHandler,
	}
}
