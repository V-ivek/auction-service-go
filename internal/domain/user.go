package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the auction system
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUser creates a new user with the given username and email
func NewUser(username, email string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the user data
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if u.Email == "" {
		return ErrInvalidEmail
	}
	return nil
}

// UpdateEmail updates the user's email
func (u *User) UpdateEmail(email string) {
	u.Email = email
	u.UpdatedAt = time.Now()
}

// UpdateUsername updates the user's username
func (u *User) UpdateUsername(username string) {
	u.Username = username
	u.UpdatedAt = time.Now()
}