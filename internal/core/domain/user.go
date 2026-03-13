package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents the authorization role assigned to a user.
type Role string

const (
	RoleAdmin       Role = "admin"
	RoleDriver      Role = "driver"
	RoleGuardian    Role = "guardian"
	RoleSchoolStaff Role = "school_staff"
)

// User represents the core user entity persisted in the database.
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone        *string   `json:"phone,omitempty" gorm:"type:varchar(20)"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;type:text;not null"`
	Role         Role      `json:"role" gorm:"type:user_role;not null;default:'guardian'"`
	FirstName    string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName     string    `json:"last_name" gorm:"type:varchar(100);not null"`
	FcmToken     *string   `json:"-" gorm:"type:text"`
	IsActive     bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the User entity.
func (User) TableName() string {
	return "users"
}

// RegisterParams holds the input parameters required to register a new user.
type RegisterParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     *string
	Role      Role
}
