package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/utils"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type authService struct {
	repo   domain.UserRepository
	rdb    *redis.Client
	secret string
}

func NewAuthService(repo domain.UserRepository, rdb *redis.Client, secret string) domain.AuthService {
	return &authService{repo: repo, rdb: rdb, secret: secret}
}

func (s *authService) Register(email, password string) error {
	hash, err := utils.Hash(password)
	if err != nil {
		return err
	}

	user := domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: hash,
		Role:     "user",
	}
	return s.repo.Create(&user)
}

func (s *authService) Login(email, password, ua, ip string) (*domain.TokenPair, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := utils.Check(user.Password, password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	at, rt, err := utils.GenerateTokenPair(user.ID, user.Role, s.secret)
	if err != nil {
		return nil, err
	}

	session := &domain.UserSession{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: rt,
		UserAgent:    ua,
		IPAddress:    ip,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}
	return &domain.TokenPair{AccessToken: at, RefreshToken: rt}, nil
}

func (s *authService) SendOTP(email string) error {
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	ctx := context.Background()

	if err := s.rdb.Set(ctx, "otp:"+email, otp, 5*time.Minute).Err(); err != nil {
		return err
	}

	go func(e, o string) {
		fmt.Printf("📧 [PRO] Sending OTP %s to %s\n", o, e)
	}(email, otp)

	return nil
}

func (s *authService) VerifyOTP(email, otp string) error {
	ctx := context.Background()
	val, err := s.rdb.Get(ctx, "otp:"+email).Result()
	if err != nil {
		return errors.New("OTP expired or invalid")
	}

	if val != otp {
		return errors.New("wrong OTP code")
	}

	user, _ := s.repo.FindByEmail(email)
	user.IsVerified = true
	s.repo.Update(user)
	s.rdb.Del(ctx, "otp:"+email)
	return nil
}

func (s *authService) RefreshToken(oldToken string) (*domain.TokenPair, error) {
	session, err := s.repo.FindSessionByToken(oldToken)
	if err != nil || time.Now().After(session.ExpiresAt) {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.FindByID(session.UserID)
	if err != nil {
		return nil, err
	}

	at, rt, _ := utils.GenerateTokenPair(user.ID, user.Role, s.secret)

	session.RefreshToken = rt
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	s.repo.UpdateSession(session)

	return &domain.TokenPair{AccessToken: at, RefreshToken: rt}, nil
}
