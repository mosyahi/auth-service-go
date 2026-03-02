package repository

import (
	"auth-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *domain.User) error { return r.db.Create(u).Error }

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.First(&u, "id = ?", id).Error
	return &u, err
}

func (r *userRepository) Update(u *domain.User) error { return r.db.Save(u).Error }

func (r *userRepository) CreateSession(s *domain.UserSession) error { return r.db.Create(s).Error }

func (r *userRepository) FindSessionByToken(t string) (*domain.UserSession, error) {
	var s domain.UserSession
	err := r.db.Where("refresh_token = ?", t).First(&s).Error
	return &s, err
}

func (r *userRepository) UpdateSession(s *domain.UserSession) error { return r.db.Save(s).Error }
