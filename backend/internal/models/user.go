package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Name      string    `json:"name"`
	Role      UserRole  `json:"role" gorm:"default:user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Forms     []Form    `json:"forms,omitempty" gorm:"foreignKey:UserID"`

	// Account lockout fields
	FailedLoginAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil         *time.Time `json:"-"`
}

// MaxLoginAttempts is the number of failed attempts before lockout
const MaxLoginAttempts = 5

// LockoutDuration is how long the account is locked after max attempts
const LockoutDuration = 15 * time.Minute

// IsLocked returns true if the account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	if time.Now().After(*u.LockedUntil) {
		return false
	}
	return true
}

// IncrementFailedAttempts increments the failed login counter and locks if needed
func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= MaxLoginAttempts {
		lockTime := time.Now().Add(LockoutDuration)
		u.LockedUntil = &lockTime
	}
}

// ResetFailedAttempts resets the failed login counter after successful login
func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New().String()
	return nil
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
