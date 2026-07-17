package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type SocialRepository interface {
	FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindFollow(ctx context.Context, followerID, followingID uuid.UUID) (*model.Follow, error)
	UpsertFollow(ctx context.Context, follow *model.Follow) error
	DeleteFollow(ctx context.Context, followerID, followingID uuid.UUID) error
	DeleteFollowsBetween(ctx context.Context, userA, userB uuid.UUID) error
	ListFollowers(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.User, int64, error)
	ListFollowing(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.User, int64, error)
	ListIncomingFollowRequests(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Follow, int64, error)
	UpdateFollowStatus(ctx context.Context, followerID, followingID uuid.UUID, status string) error
	IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	FindBlock(ctx context.Context, blockerID, blockedID uuid.UUID) (*model.Block, error)
	FindBlocksByBlocker(ctx context.Context, blockerID uuid.UUID) ([]model.Block, error)
	CreateBlock(ctx context.Context, block *model.Block) error
	DeleteBlock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	CreateReport(ctx context.Context, report *model.Report) error
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
	PostExists(ctx context.Context, id uuid.UUID) (bool, error)
	CommentExists(ctx context.Context, id uuid.UUID) (bool, error)
	CreateNotification(ctx context.Context, notification *model.Notification) error
	// Close Friends
	SetCloseFriend(ctx context.Context, followerID, followingID uuid.UUID, isCloseFriend bool) error
	ListCloseFriends(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.User, int64, error)
	IsCloseFriend(ctx context.Context, userID, targetID uuid.UUID) (bool, error)
}

type GormSocialRepository struct {
	db *gorm.DB
}

func NewSocialRepository(db *gorm.DB) SocialRepository {
	return &GormSocialRepository{db: db}
}

func (r *GormSocialRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormSocialRepository) FindFollow(ctx context.Context, followerID, followingID uuid.UUID) (*model.Follow, error) {
	var follow model.Follow
	err := r.db.WithContext(ctx).First(&follow, "follower_id = ? AND following_id = ?", followerID, followingID).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *GormSocialRepository) UpsertFollow(ctx context.Context, follow *model.Follow) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "follower_id"}, {Name: "following_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(follow).Error
}

func (r *GormSocialRepository) DeleteFollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.Follow{}).Error
}

func (r *GormSocialRepository) DeleteFollowsBetween(ctx context.Context, userA, userB uuid.UUID) error {
	return r.db.WithContext(ctx).Where("(follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)", userA, userB, userB, userA).Delete(&model.Follow{}).Error
}

func (r *GormSocialRepository) ListFollowers(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.User{}).Joins("JOIN follows ON follows.follower_id = users.id").Where("follows.following_id = ? AND follows.status = ?", userID, "accepted")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := query.Order("follows.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *GormSocialRepository) ListFollowing(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.User{}).Joins("JOIN follows ON follows.following_id = users.id").Where("follows.follower_id = ? AND follows.status = ?", userID, "accepted")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := query.Order("follows.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *GormSocialRepository) ListIncomingFollowRequests(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Follow, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Follow{}).Preload("Follower").Where("following_id = ? AND status = ?", userID, "pending")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var follows []model.Follow
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&follows).Error
	return follows, total, err
}

func (r *GormSocialRepository) UpdateFollowStatus(ctx context.Context, followerID, followingID uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormSocialRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	return CreateNotificationHelper(ctx, r.db, notification)
}

func (r *GormSocialRepository) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Block{}).Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userA, userB, userB, userA).Count(&count).Error
	return count > 0, err
}

func (r *GormSocialRepository) FindBlock(ctx context.Context, blockerID, blockedID uuid.UUID) (*model.Block, error) {
	var block model.Block
	err := r.db.WithContext(ctx).First(&block, "blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *GormSocialRepository) CreateBlock(ctx context.Context, block *model.Block) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(block).Error
}

func (r *GormSocialRepository) DeleteBlock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&model.Block{}).Error
}

func (r *GormSocialRepository) FindBlocksByBlocker(ctx context.Context, blockerID uuid.UUID) ([]model.Block, error) {
	var blocks []model.Block
	err := r.db.WithContext(ctx).Where("blocker_id = ?", blockerID).Preload("Blocked").Find(&blocks).Error
	return blocks, err
}

func (r *GormSocialRepository) CreateReport(ctx context.Context, report *model.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *GormSocialRepository) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *GormSocialRepository) PostExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormSocialRepository) CommentExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

// Close Friends implementation

func (r *GormSocialRepository) SetCloseFriend(ctx context.Context, followerID, followingID uuid.UUID, isCloseFriend bool) error {
	var follow model.Follow
	err := r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").First(&follow).Error
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Update("is_close_friend", isCloseFriend).Error
}

func (r *GormSocialRepository) ListCloseFriends(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.User{}).
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.following_id = ? AND follows.status = ? AND follows.is_close_friend = ?", userID, "accepted", true)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := query.Order("follows.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *GormSocialRepository) IsCloseFriend(ctx context.Context, userID, targetID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ? AND status = ? AND is_close_friend = ?", userID, targetID, "accepted", true).
		Count(&count).Error
	return count > 0, err
}
