package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
)

type AuthRepository interface {
	Register(req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
	VerifyOTP(req dto.VerifyOTPRequest) (*dto.AuthResponse, error)
	ForgotPassword(req dto.ForgotPasswordRequest) error
	RecoverUsername(req dto.RecoverUsernameRequest) error
	RecoverEmail(req dto.RecoverEmailRequest) error
	ResetPassword(req dto.ResetPasswordRequest) error
	IsUsernameAvailable(username string) (bool, error)
	IsEmailAvailable(email string) (bool, error)
}

type SupabaseAuthRepository struct {
	baseURL    string
	authKey    string
	httpClient *http.Client
	db         *gorm.DB
}

func NewAuthRepository(db *gorm.DB, baseURL, authKey string) AuthRepository {
	return &SupabaseAuthRepository{db: db, baseURL: strings.TrimRight(baseURL, "/"), authKey: authKey, httpClient: &http.Client{}}
}

func (r *SupabaseAuthRepository) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	payload := map[string]any{"email": req.Email, "password": req.Password, "data": map[string]any{"name": req.Name, "username": req.Username, "phone": req.Phone}}
	var out map[string]any
	if err := r.post("/auth/v1/signup", payload, &out); err != nil { return nil, err }
	return buildAuthResponse(out), nil
}

func (r *SupabaseAuthRepository) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	payload := map[string]any{"email": req.Identifier, "password": req.Password}
	var out map[string]any
	if err := r.post("/auth/v1/token?grant_type=password", payload, &out); err != nil { return nil, err }
	return buildAuthResponse(out), nil
}

func (r *SupabaseAuthRepository) VerifyOTP(req dto.VerifyOTPRequest) (*dto.AuthResponse, error) {
	payload := map[string]any{"email": req.Email, "token": req.Token, "type": req.Type}
	var out map[string]any
	if err := r.post("/auth/v1/verify", payload, &out); err != nil { return nil, err }
	return buildAuthResponse(out), nil
}

func (r *SupabaseAuthRepository) ForgotPassword(req dto.ForgotPasswordRequest) error {
	payload := map[string]any{"email": req.Email}
	var out map[string]any
	return r.post("/auth/v1/recover", payload, &out)
}

func (r *SupabaseAuthRepository) RecoverUsername(req dto.RecoverUsernameRequest) error {
	payload := map[string]any{"email": req.Email}
	var out map[string]any
	return r.post("/auth/v1/recover", payload, &out)
}

func (r *SupabaseAuthRepository) RecoverEmail(req dto.RecoverEmailRequest) error {
	payload := map[string]any{"username": req.Username, "phone": req.Phone}
	var out map[string]any
	return r.post("/auth/v1/recover-email", payload, &out)
}

func (r *SupabaseAuthRepository) ResetPassword(req dto.ResetPasswordRequest) error {
	payload := map[string]any{"password": req.Password}
	var out map[string]any
	return r.post("/auth/v1/user", payload, &out)
}

func (r *SupabaseAuthRepository) IsUsernameAvailable(username string) (bool, error) {
	var count int64
	err := r.db.Table("users").Where("username = ?", username).Count(&count).Error
	return count == 0, err
}

func (r *SupabaseAuthRepository) IsEmailAvailable(email string) (bool, error) {
	var count int64
	err := r.db.Table("users").Where("email = ?", email).Count(&count).Error
	return count == 0, err
}

func (r *SupabaseAuthRepository) post(path string, payload any, out any) error {
	if r.authKey == "" {
		return fmt.Errorf("supabase auth error: SUPABASE_ANON_KEY is not configured")
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, r.baseURL+path, bytes.NewReader(body))
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	if r.authKey != "" {
		req.Header.Set("apikey", r.authKey)
		req.Header.Set("Authorization", "Bearer "+r.authKey)
	}
	resp, err := r.httpClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		msg := resp.Status
		if msgDetail, ok := errResp["msg"].(string); ok { msg = msgDetail } else if msgDetail, ok := errResp["message"].(string); ok { msg = msgDetail }
		return fmt.Errorf("supabase auth error: %s", msg)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func buildAuthResponse(out map[string]any) *dto.AuthResponse {
	resp := &dto.AuthResponse{Message: "Success"}
	if user, ok := out["user"].(map[string]any); ok {
		resp.User = dto.AuthUserResponse{ID: fmt.Sprint(user["id"]), Email: fmt.Sprint(user["email"])}
	} else if id, ok := out["id"]; ok {
		resp.User = dto.AuthUserResponse{ID: fmt.Sprint(id), Email: fmt.Sprint(out["email"])}
	}

	if session, ok := out["session"].(map[string]any); ok {
		resp.Session = &dto.AuthSessionResponse{AccessToken: fmt.Sprint(session["access_token"]), RefreshToken: fmt.Sprint(session["refresh_token"]), TokenType: fmt.Sprint(session["token_type"])}
	} else if at, ok := out["access_token"]; ok {
		resp.Session = &dto.AuthSessionResponse{AccessToken: fmt.Sprint(at), RefreshToken: fmt.Sprint(out["refresh_token"]), TokenType: fmt.Sprint(out["token_type"])}
	}

	if resp.User.ID == "<nil>" { resp.User.ID = "" }
	if resp.User.Email == "<nil>" { resp.User.Email = "" }

	return resp
}


