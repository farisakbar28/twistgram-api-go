package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type UserRepository interface {
	FindByID(id uuid.UUID) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	UsernameExists(username string, excludeID uuid.UUID) (bool, error)
	Update(user *model.User) error
	CountFollowers(userID uuid.UUID) (int64, error)
	CountFollowing(userID uuid.UUID) (int64, error)
	CountPosts(userID uuid.UUID) (int64, error)
	IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error)
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) UsernameExists(username string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Where("username = ? AND id <> ?", username, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *GormUserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *GormUserRepository) CountFollowers(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("following_id = ? AND status = ?", userID, "accepted").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountFollowing(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND status = ?", userID, "accepted").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountPosts(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Post{}).
		Where("user_id = ? AND deleted_at IS NULL AND is_archived = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error) {
	if followerID == uuid.Nil || followingID == uuid.Nil {
		return false, nil
	}

	var follow model.Follow
	err := r.db.First(&follow, "follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
