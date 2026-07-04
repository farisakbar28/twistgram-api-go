package service

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/pkg/auth"
	"twistgram-api-go/pkg/mailer"
)

var (
	ErrAuthUnavailable = errors.New("auth unavailable")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	repo repository.AuthRepository
	cfg  *config.Config
}

func NewAuthService(repo repository.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" || req.Username == "" || req.Password == "" || req.Name == "" { return nil, ErrInvalidInput }
	if !isValidPassword(req.Password) { return nil, errors.New("password must be at least 8 chars, contain an uppercase letter and a number/symbol") }

	// Cek ketersediaan email
	existingUser, err := s.repo.FindUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		if !existingUser.EmailVerified {
			return nil, errors.New("email already registered but not verified. please request a new OTP")
		}
		return nil, errors.New("email already registered")
	}

	userAvail, err := s.repo.IsUsernameAvailable(req.Username)
	if err != nil { return nil, err }
	if !userAvail { return nil, errors.New("username already taken") }

	hash, err := auth.HashPassword(req.Password)
	if err != nil { return nil, err }

	var phone *string
	if strings.TrimSpace(req.Phone) != "" { p := strings.TrimSpace(req.Phone); phone = &p }

	user := &model.User{
		Name:          req.Name,
		Username:      req.Username,
		Email:         req.Email,
		Phone:         phone,
		PasswordHash:  &hash,
		EmailVerified: false,
	}

	otpCode, _ := auth.GenerateOTP()
	otp := &model.AuthOTP{
		Code:      otpCode,
		Type:      "signup",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.repo.CreateUserWithOTP(user, otp); err != nil { return nil, err }
	_ = mailer.SendOTPEmail(user.Email, otpCode)

	return &dto.AuthResponse{
		Message: "User registered, please verify OTP",
		User:    dto.AuthUserResponse{ID: user.ID.String(), Email: user.Email},
	}, nil
}

func isValidPassword(p string) bool {
	if len(p) < 8 { return false }
	var hasUpper, hasSpecialOrNumber bool
	for i, c := range p {
		if i == 0 && unicode.IsUpper(c) { hasUpper = true }
		if unicode.IsNumber(c) || unicode.IsPunct(c) || unicode.IsSymbol(c) { hasSpecialOrNumber = true }
	}
	return hasUpper && hasSpecialOrNumber
}

func (s *AuthService) ResendOTP(req dto.ResendOTPRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Type == "" { return ErrInvalidInput }

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil || user == nil { return nil } // Silent return to prevent email enumeration

	if req.Type == "signup" && user.EmailVerified {
		return errors.New("email is already verified")
	}

	// Delete old OTPs for this type
	_ = s.repo.DeleteOTPByUserID(user.ID, req.Type)

	otpCode, _ := auth.GenerateOTP()
	otp := &model.AuthOTP{
		UserID:    user.ID,
		Code:      otpCode,
		Type:      req.Type,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	_ = s.repo.SaveOTP(otp)
	_ = mailer.SendOTPEmail(user.Email, otpCode)

	return nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	req.Identifier = strings.TrimSpace(strings.ToLower(req.Identifier))
	if req.Identifier == "" || req.Password == "" { return nil, ErrInvalidInput }

	var user *model.User
	var err error

	if strings.Contains(req.Identifier, "@") {
		user, err = s.repo.FindUserByEmail(req.Identifier)
	} else {
		user, err = s.repo.FindUserByUsername(req.Identifier)
	}

	if err != nil || user == nil { return nil, ErrInvalidCredentials }

	if user.PasswordHash == nil || !auth.CheckPasswordHash(req.Password, *user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return nil, errors.New("Email not confirmed")
	}

	accessToken, refreshToken, err := auth.GenerateJWT(user.ID, user.Email, s.cfg.SupabaseJWTSecret)
	if err != nil { return nil, err }

	return &dto.AuthResponse{
		Message: "Success",
		User:    dto.AuthUserResponse{ID: user.ID.String(), Email: user.Email},
		Session: &dto.AuthSessionResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
		},
	}, nil
}

func (s *AuthService) VerifyOTP(req dto.VerifyOTPRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Token = strings.TrimSpace(req.Token)
	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	if req.Email == "" || req.Token == "" || req.Type == "" { return nil, ErrInvalidInput }

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil || user == nil { return nil, ErrInvalidInput }

	otp, err := s.repo.FindValidOTP(user.ID, req.Token, req.Type)
	if err != nil || otp == nil { return nil, errors.New("Token has expired or is invalid") }

	_ = s.repo.DeleteOTP(otp.ID)

	if req.Type == "signup" {
		user.EmailVerified = true
		_ = s.repo.UpdateUser(user)
	}

	accessToken, refreshToken, err := auth.GenerateJWT(user.ID, user.Email, s.cfg.SupabaseJWTSecret)
	if err != nil { return nil, err }

	return &dto.AuthResponse{
		Message: "Success",
		User:    dto.AuthUserResponse{ID: user.ID.String(), Email: user.Email},
		Session: &dto.AuthSessionResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
		},
	}, nil
}

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" { return ErrInvalidInput }

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil || user == nil { return nil } // Jangan bocorkan error kalau email tidak ada

	otpCode, _ := auth.GenerateOTP()
	otp := &model.AuthOTP{
		UserID:    user.ID,
		Code:      otpCode,
		Type:      "recovery",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	_ = s.repo.SaveOTP(otp)
	_ = mailer.SendOTPEmail(user.Email, otpCode)

	return nil
}

func (s *AuthService) RecoverUsername(req dto.RecoverUsernameRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" { return ErrInvalidInput }
	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil || user == nil { return nil }

	// Skenario A (Lupa Username) - Kirim username ke email
	subject := "Pemulihan Username Twistgram"
	body := "Halo,\nUsername Twistgram Anda adalah: " + user.Username
	_ = mailer.SendEmail(user.Email, subject, body)
	return nil
}

func (s *AuthService) RecoverEmail(req dto.RecoverEmailRequest) error {
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Username == "" || req.Phone == "" { return ErrInvalidInput }
	return nil
}

func (s *AuthService) CheckAvailability(req dto.CheckAvailabilityRequest) (*dto.CheckAvailabilityResponse, error) {
	username := strings.TrimSpace(strings.ToLower(req.Username))
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if username == "" && email == "" { return nil, ErrInvalidInput }
	if username != "" {
		available, err := s.repo.IsUsernameAvailable(username)
		if err != nil { return nil, err }
		if !available { return &dto.CheckAvailabilityResponse{Available: false, Message: "username already taken"}, nil }
	}
	if email != "" {
		available, err := s.repo.IsEmailAvailable(email)
		if err != nil { return nil, err }
		if !available { return &dto.CheckAvailabilityResponse{Available: false, Message: "email already registered"}, nil }
	}
	return &dto.CheckAvailabilityResponse{Available: true}, nil
}

func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Token = strings.TrimSpace(req.Token)
	if req.Email == "" || req.Token == "" || req.Password == "" { return ErrInvalidInput }
	if !isValidPassword(req.Password) { return errors.New("password must be at least 8 chars, contain an uppercase letter and a number/symbol") }

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil || user == nil { return errors.New("invalid recovery token") }

	otp, err := s.repo.FindValidOTP(user.ID, req.Token, "recovery")
	if err != nil || otp == nil { return errors.New("invalid recovery token") }

	hash, err := auth.HashPassword(req.Password)
	if err != nil { return err }

	user.PasswordHash = &hash
	if err := s.repo.UpdateUser(user); err != nil { return err }

	_ = s.repo.DeleteOTP(otp.ID)

	return nil
}

func (s *AuthService) RefreshToken(req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" { return nil, ErrInvalidInput }

	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.SupabaseJWTSecret), nil
	})

	if err != nil || !token.Valid { return nil, errors.New("invalid refresh token") }

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" { return nil, errors.New("invalid token type") }

	sub, ok := claims["sub"].(string)
	if !ok { return nil, errors.New("invalid token payload") }

	userID, err := uuid.Parse(sub)
	if err != nil { return nil, errors.New("invalid user id in token") }

	user, err := s.repo.FindUserByID(userID)
	if err != nil || user == nil { return nil, errors.New("user not found") }

	accessToken, refreshToken, err := auth.GenerateJWT(user.ID, user.Email, s.cfg.SupabaseJWTSecret)
	if err != nil { return nil, err }

	return &dto.AuthResponse{
		Message: "Token refreshed successfully",
		User:    dto.AuthUserResponse{ID: user.ID.String(), Email: user.Email},
		Session: &dto.AuthSessionResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
		},
	}, nil
}
