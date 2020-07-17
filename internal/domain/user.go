package domain

import (
	"context"
	"time"
)

// User models
type User struct {
	UUID        string    `json:"uuid" db:"uuid"`
	Email       string    `json:"email" db:"email" validate:"required,email"`
	Password    string    `json:"password" db:"password" validate:"required"`
	NewPassword *string   `json:"new_password" db:"new_password"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Salt        string    `json:"salt" db:"salt"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserRepository represent the users's repository contract
type UserRepository interface {
	Find(ctx context.Context, uuid string) (*User, error)
	FindOneBy(ctx context.Context, criteria map[string]interface{}, orderBy *map[string]string) (*User, error)
	FindAll(context.Context) ([]*User, error)
	FindBy(ctx context.Context, criteria map[string]interface{}, orderBy *map[string]string, limit *uint, offest *uint) ([]*User, error)
	Store(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
}

// UserUsecase represent the users's usecase contract
type UserUsecase interface {
	Register(ctx context.Context, user *User) (token string, err error)
	Login(ctx context.Context, user *User) (token string, err error)
	ChangeEmail(ctx context.Context, user *User, token JWToken) error
	Activation(ctx context.Context, token JWToken) error
	ChangePassword(ctx context.Context, user *User, token JWToken) (tokenConfirmation string, err error)
	PasswordConfirm(ctx context.Context, token JWToken) error
	ForgotPasswordRequest(ctx context.Context, email string) (token string, err error)
	ForgotPasswordConfirm(ctx context.Context, user *User, token JWToken) error
}
