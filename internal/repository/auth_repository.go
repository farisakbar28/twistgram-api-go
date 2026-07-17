package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type AuthRepository interface {
	CreateUserWithOTP(ctx context.Context, user *model.User, otp *model.AuthOTP) error
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
	FindUserByPhone(ctx context.Context, phone string) (*model.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	SaveOTP(ctx context.Context, otp *model.AuthOTP) error
	FindValidOTP(ctx context.Context, userID uuid.UUID, code, otpType string) (*model.AuthOTP, error)
	DeleteOTP(ctx context.Context, id uuid.UUID) error
	DeleteOTPByUserID(ctx context.Context, userID uuid.UUID, otpType string) error
	UpdateUser(ctx context.Context, user *model.User) error
	IsUsernameAvailable(ctx context.Context, username string) (bool, error)
	IsEmailAvailable(ctx context.Context, email string) (bool, error)
}

type LocalAuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &LocalAuthRepository{db: db}
}

func (r *LocalAuthRepository) CreateUserWithOTP(ctx context.Context, user *model.User, otp *model.AuthOTP) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		otp.UserID = user.ID
		return tx.Create(otp).Error
	})
}

func (r *LocalAuthRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) SaveOTP(ctx context.Context, otp *model.AuthOTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *LocalAuthRepository) FindValidOTP(ctx context.Context, userID uuid.UUID, code, otpType string) (*model.AuthOTP, error) {
	var otp model.AuthOTP
	err := r.db.WithContext(ctx).Where("user_id = ? AND code = ? AND type = ? AND expires_at > ?", userID, code, otpType, time.Now()).First(&otp).Error
	return &otp, err
}

func (r *LocalAuthRepository) DeleteOTP(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AuthOTP{}).Error
}

func (r *LocalAuthRepository) DeleteOTPByUserID(ctx context.Context, userID uuid.UUID, otpType string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND type = ?", userID, otpType).Delete(&model.AuthOTP{}).Error
}

func (r *LocalAuthRepository) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *LocalAuthRepository) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("users").Where("username = ?", username).Count(&count).Error
	return count == 0, err
}

func (r *LocalAuthRepository) IsEmailAvailable(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("users").Where("email = ?", email).Count(&count).Error
	return count == 0, err
}
