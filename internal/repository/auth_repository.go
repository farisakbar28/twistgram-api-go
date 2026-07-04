package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type AuthRepository interface {
	CreateUserWithOTP(user *model.User, otp *model.AuthOTP) error
	FindUserByEmail(email string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	FindUserByPhone(phone string) (*model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
	SaveOTP(otp *model.AuthOTP) error
	FindValidOTP(userID uuid.UUID, code, otpType string) (*model.AuthOTP, error)
	DeleteOTP(id uuid.UUID) error
	DeleteOTPByUserID(userID uuid.UUID, otpType string) error
	UpdateUser(user *model.User) error
	IsUsernameAvailable(username string) (bool, error)
	IsEmailAvailable(email string) (bool, error)
}

type LocalAuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &LocalAuthRepository{db: db}
}

func (r *LocalAuthRepository) CreateUserWithOTP(user *model.User, otp *model.AuthOTP) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil { return err }
		otp.UserID = user.ID
		return tx.Create(otp).Error
	})
}

func (r *LocalAuthRepository) FindUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) FindUserByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) FindUserByPhone(phone string) (*model.User, error) {
	var u model.User
	err := r.db.Where("phone = ?", phone).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) FindUserByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.Where("id = ?", id).First(&u).Error
	return &u, err
}

func (r *LocalAuthRepository) SaveOTP(otp *model.AuthOTP) error {
	return r.db.Create(otp).Error
}

func (r *LocalAuthRepository) FindValidOTP(userID uuid.UUID, code, otpType string) (*model.AuthOTP, error) {
	var otp model.AuthOTP
	err := r.db.Where("user_id = ? AND code = ? AND type = ? AND expires_at > ?", userID, code, otpType, time.Now()).First(&otp).Error
	return &otp, err
}

func (r *LocalAuthRepository) DeleteOTP(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.AuthOTP{}).Error
}

func (r *LocalAuthRepository) DeleteOTPByUserID(userID uuid.UUID, otpType string) error {
	return r.db.Where("user_id = ? AND type = ?", userID, otpType).Delete(&model.AuthOTP{}).Error
}

func (r *LocalAuthRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *LocalAuthRepository) IsUsernameAvailable(username string) (bool, error) {
	var count int64
	err := r.db.Table("users").Where("username = ?", username).Count(&count).Error
	return count == 0, err
}

func (r *LocalAuthRepository) IsEmailAvailable(email string) (bool, error) {
	var count int64
	err := r.db.Table("users").Where("email = ?", email).Count(&count).Error
	return count == 0, err
}
