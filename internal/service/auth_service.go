package service

import (
	"errors"
	"strings"

	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/repository"
)

var ErrAuthUnavailable = errors.New("auth unavailable")

type AuthService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) *AuthService { return &AuthService{repo: repo} }

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" || req.Username == "" || req.Password == "" || req.Name == "" { return nil, ErrInvalidInput }
	return s.repo.Register(req)
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	req.Identifier = strings.TrimSpace(strings.ToLower(req.Identifier))
	if req.Identifier == "" || req.Password == "" { return nil, ErrInvalidInput }
	return s.repo.Login(req)
}

func (s *AuthService) VerifyOTP(req dto.VerifyOTPRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Token = strings.TrimSpace(req.Token)
	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	if req.Email == "" || req.Token == "" || req.Type == "" { return nil, ErrInvalidInput }
	return s.repo.VerifyOTP(req)
}

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" { return ErrInvalidInput }
	return s.repo.ForgotPassword(req)
}

func (s *AuthService) RecoverUsername(req dto.RecoverUsernameRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" { return ErrInvalidInput }
	return s.repo.RecoverUsername(req)
}

func (s *AuthService) RecoverEmail(req dto.RecoverEmailRequest) error {
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Username == "" || req.Phone == "" { return ErrInvalidInput }
	return s.repo.RecoverEmail(req)
}

func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" || req.Password == "" { return ErrInvalidInput }
	return s.repo.ResetPassword(req)
}
