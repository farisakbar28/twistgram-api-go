package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	UsernameExists(ctx context.Context, username string, excludeID uuid.UUID) (bool, error)
	Update(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	IncrementTokenVersion(ctx context.Context, id uuid.UUID) error
	CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error)
	CountPosts(ctx context.Context, userID uuid.UUID) (int64, error)
	GetInterests(ctx context.Context, userID uuid.UUID) ([]string, error)
	SetInterests(ctx context.Context, userID uuid.UUID, interests []string) error
	IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error)
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) UsernameExists(ctx context.Context, username string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("username = ? AND id <> ?", username, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *GormUserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *GormUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clean up related data
		_ = tx.Where("user_id = ?", id).Delete(&model.UserInterest{}).Error
		_ = tx.Where("follower_id = ? OR following_id = ?", id, id).Delete(&model.Follow{}).Error
		_ = tx.Where("blocker_id = ? OR blocked_id = ?", id, id).Delete(&model.Block{}).Error
		_ = tx.Where("sender_id = ?", id).Delete(&model.Message{}).Error
		_ = tx.Where("user_id = ?", id).Delete(&model.ConversationParticipant{}).Error
		_ = tx.Where("recipient_id = ? OR actor_id = ?", id, id).Delete(&model.Notification{}).Error
		_ = tx.Where("reporter_id = ?", id).Delete(&model.Report{}).Error
		_ = tx.Where("user_id = ?", id).Delete(&model.AuthOTP{}).Error
		// Soft delete posts
		_ = tx.Model(&model.Post{}).Where("user_id = ?", id).Update("deleted_at", gorm.Expr("now()")).Error
		// Hard delete user
		return tx.Delete(&model.User{}, "id = ?", id).Error
	})
}

func (r *GormUserRepository) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("token_version", gorm.Expr("token_version + 1")).Error
}

func (r *GormUserRepository) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("following_id = ? AND status = ?", userID, "accepted").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND status = ?", userID, "accepted").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountPosts(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Post{}).
		Where("user_id = ? AND deleted_at IS NULL AND is_archived = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) GetInterests(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var interests []model.UserInterest
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&interests).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(interests))
	for _, i := range interests {
		out = append(out, i.InterestCategory)
	}
	return out, nil
}

func (r *GormUserRepository) SetInterests(ctx context.Context, userID uuid.UUID, interests []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserInterest{}).Error; err != nil {
			return err
		}
		if len(interests) == 0 {
			return nil
		}
		items := make([]model.UserInterest, 0, len(interests))
		for _, cat := range interests {
			items = append(items, model.UserInterest{UserID: userID, InterestCategory: cat})
		}
		return tx.Create(&items).Error
	})
}

func (r *GormUserRepository) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	if followerID == uuid.Nil || followingID == uuid.Nil {
		return false, nil
	}

	var follow model.Follow
	err := r.db.WithContext(ctx).First(&follow, "follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *GormUserRepository) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	if userA == uuid.Nil || userB == uuid.Nil || userA == userB {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Block{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userA, userB, userB, userA).
		Count(&count).Error
	return count > 0, err
}
