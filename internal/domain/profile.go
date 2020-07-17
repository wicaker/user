package domain

import (
	"context"
	"time"
)

// Profile models
type Profile struct {
	UUID      string     `json:"uuid" db:"uuid"`
	User      User       `json:"user" db:"user"`
	UserUUID  string     `json:"user_uuid" db:"user_uuid"`
	FirstName *string    `json:"first_name" db:"first_name"`
	LastName  *string    `json:"last_name" db:"last_name"`
	Phone     *string    `json:"phone" db:"phone"`
	Address   *string    `json:"address" db:"address"`
	Gender    *string    `json:"gender" db:"gender"`
	Dob       *time.Time `json:"dob" db:"dob"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// ProfileRepository represent profile's repository contract
type ProfileRepository interface {
	Find(ctx context.Context, uuid string) (*Profile, error)
	FindOneBy(ctx context.Context, criteria map[string]interface{}, orderBy *map[string]string) (*Profile, error)
	FindAll(context.Context) ([]*Profile, error)
	FindBy(ctx context.Context, criteria map[string]interface{}, orderBy *map[string]string, limit *uint, offest *uint) ([]*Profile, error)
	Store(ctx context.Context, profile *Profile) (*Profile, error)
	Update(ctx context.Context, profile *Profile) error
}

// ProfileUsecase represent profile's usecase contract
type ProfileUsecase interface {
	Store(ctx context.Context, profile *Profile) (*Profile, error)
	GetByUUID(ctx context.Context, uuid string) (*Profile, error)
	Fetch(context.Context) ([]*Profile, error)
	Update(ctx context.Context, profile *Profile) (*Profile, error)
}
