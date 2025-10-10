package main

import (
	"fmt"
	"net/http"

	"github.com/kivaplus/kivaplus-backend/lambda/app"
	"github.com/kivaplus/kivaplus-backend/lambda/middleware"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Username string `json:"username"`
}

func HandleRequest(event MyEvent) (string, error) {
	if event.Username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	return fmt.Sprintf("Successfully called by - %s", event.Username), nil
}

func ProtectedRoute(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "It's works! Nobody can see if not logged.",
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambdaApp := app.NewApp()
	lambda.Start(func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		switch request.Path {
		case "/register":
			return lambdaApp.ApiHandler.RegisterUserHandler(request)
		case "/login":
			return lambdaApp.ApiHandler.LoginUserHandler(request)
		case "/protected":
			return middleware.ValidatePasswordJWTMiddleware(ProtectedRoute)(request)
		default:
			return events.APIGatewayProxyResponse{
				Body:       "Not found",
				StatusCode: http.StatusNotFound,
			}, nil
		}
	})
}
