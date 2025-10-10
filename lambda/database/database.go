package database

import (
	"context"
	"fmt"
	"os"

	"github.com/kivaplus/kivaplus-backend/lambda/types"

	// "kivaplus/backend/lambda/types"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserStore interface {
	DoesUserExist(username string) (bool, error)
	InsertUser(user types.User) error
	GetUser(username string) (types.User, error)
}

type DynamoDBClient struct {
	databaseStore *dynamodb.Client
}

var TABLE_NAME = os.Getenv("USER_TABLE_NAME")

func NewDynamoDBClient() DynamoDBClient {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	db := dynamodb.NewFromConfig(cfg)
	return DynamoDBClient{
		databaseStore: db,
	}
}

func (u DynamoDBClient) DoesUserExist(username string) (bool, error) {
	res, err := u.databaseStore.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]dynamoTypes.AttributeValue{
			"username": &dynamoTypes.AttributeValueMemberS{Value: username},
		},
	})
	if err != nil {
		return true, err
	}
	if res.Item == nil {
		return false, nil
	}

	return true, nil
}

func (u DynamoDBClient) InsertUser(user types.User) error {
	//assemble the item
	item := &dynamodb.PutItemInput{
		TableName: aws.String(TABLE_NAME),
		Item: map[string]dynamoTypes.AttributeValue{
			"username": &dynamoTypes.AttributeValueMemberS{Value: user.Username},
			"password": &dynamoTypes.AttributeValueMemberS{Value: user.PasswordHash},
		},
	}

	_, err := u.databaseStore.PutItem(context.TODO(), item)
	if err != nil {
		return err
	}

	return nil
}

func (u DynamoDBClient) GetUser(username string) (types.User, error) {
	var user types.User

	result, err := u.databaseStore.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]dynamoTypes.AttributeValue{
			"username": &dynamoTypes.AttributeValueMemberS{Value: username},
		},
	})

	if err != nil {
		return user, err
	}

	if result.Item == nil {
		return user, fmt.Errorf("user not found")
	}

	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
