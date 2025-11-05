package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose password hash in JSON
	Active       bool       `json:"active"`
	LastAccess   *time.Time `json:"last_access,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// NewUser creates a new user with hashed password
func NewUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidatePassword checks if the provided password matches the user's password
func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdateLastAccess updates the user's last access time
func (u *User) UpdateLastAccess() {
	now := time.Now()
	u.LastAccess = &now
	u.UpdatedAt = now
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}

// Activate activates the user account
func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Active && u.DeletedAt == nil
}
