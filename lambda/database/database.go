package database

import (
	"github.com/kivaplus/kivaplus-backend/lambda/types"

	"log"
)

type UserStore interface {
	DoesUserExist(email string) (bool, error)
	InsertUser(user types.User) error
	GetUser(email string) (types.User, error)
}

// NewUserStore creates a UserStore based on environment configuration
func NewUserStore() (UserStore, error) {
	log.Println("🐘 Using PostgreSQL database")
	return NewPostgresClient()
}
