package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email      string    `gorm:"uniqueIndex;not null"`
	Password   string    `gorm:"not null"`
	Role       string    `gorm:"type:varchar(20);default:'user'"`
	IsVerified bool      `gorm:"default:false"`
	CreatedAt  time.Time
}

type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;index"`
	RefreshToken string    `gorm:"uniqueIndex"`
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id uuid.UUID) (*User, error)
	Update(user *User) error
	CreateSession(session *UserSession) error
	FindSessionByToken(token string) (*UserSession, error)
	UpdateSession(session *UserSession) error
}

type AuthService interface {
	Register(email, password string) error
	Login(email, password, ua, ip string) (*TokenPair, error)
	RefreshToken(oldToken string) (*TokenPair, error)
	SendOTP(email string) error
	VerifyOTP(email, otp string) error
}
