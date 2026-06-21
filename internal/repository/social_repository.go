package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type SocialRepository interface {
	FindUserByID(id uuid.UUID) (*model.User, error)
	FindFollow(followerID, followingID uuid.UUID) (*model.Follow, error)
	UpsertFollow(follow *model.Follow) error
	DeleteFollow(followerID, followingID uuid.UUID) error
	DeleteFollowsBetween(userA, userB uuid.UUID) error
	ListFollowers(userID uuid.UUID, page, limit int) ([]model.User, int64, error)
	ListFollowing(userID uuid.UUID, page, limit int) ([]model.User, int64, error)
	ListIncomingFollowRequests(userID uuid.UUID, page, limit int) ([]model.Follow, int64, error)
	UpdateFollowStatus(followerID, followingID uuid.UUID, status string) error
	IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error)
	FindBlock(blockerID, blockedID uuid.UUID) (*model.Block, error)
	CreateBlock(block *model.Block) error
	DeleteBlock(blockerID, blockedID uuid.UUID) error
	CreateReport(report *model.Report) error
	UserExists(id uuid.UUID) (bool, error)
	PostExists(id uuid.UUID) (bool, error)
	CommentExists(id uuid.UUID) (bool, error)
}

type GormSocialRepository struct {
	db *gorm.DB
}

func NewSocialRepository(db *gorm.DB) SocialRepository {
	return &GormSocialRepository{db: db}
}

func (r *GormSocialRepository) FindUserByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormSocialRepository) FindFollow(followerID, followingID uuid.UUID) (*model.Follow, error) {
	var follow model.Follow
	err := r.db.First(&follow, "follower_id = ? AND following_id = ?", followerID, followingID).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *GormSocialRepository) UpsertFollow(follow *model.Follow) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "follower_id"}, {Name: "following_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(follow).Error
}

func (r *GormSocialRepository) DeleteFollow(followerID, followingID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.Follow{}).Error
}

func (r *GormSocialRepository) DeleteFollowsBetween(userA, userB uuid.UUID) error {
	return r.db.Where("(follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)", userA, userB, userB, userA).Delete(&model.Follow{}).Error
}

func (r *GormSocialRepository) ListFollowers(userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	var total int64
	query := r.db.Model(&model.User{}).Joins("JOIN follows ON follows.follower_id = users.id").Where("follows.following_id = ? AND follows.status = ?", userID, "accepted")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := query.Order("follows.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *GormSocialRepository) ListFollowing(userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	var total int64
	query := r.db.Model(&model.User{}).Joins("JOIN follows ON follows.following_id = users.id").Where("follows.follower_id = ? AND follows.status = ?", userID, "accepted")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := query.Order("follows.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *GormSocialRepository) ListIncomingFollowRequests(userID uuid.UUID, page, limit int) ([]model.Follow, int64, error) {
	var total int64
	query := r.db.Model(&model.Follow{}).Preload("Follower").Where("following_id = ? AND status = ?", userID, "pending")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var follows []model.Follow
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&follows).Error
	return follows, total, err
}

func (r *GormSocialRepository) UpdateFollowStatus(followerID, followingID uuid.UUID, status string) error {
	result := r.db.Model(&model.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormSocialRepository) IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Block{}).Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userA, userB, userB, userA).Count(&count).Error
	return count > 0, err
}

func (r *GormSocialRepository) FindBlock(blockerID, blockedID uuid.UUID) (*model.Block, error) {
	var block model.Block
	err := r.db.First(&block, "blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *GormSocialRepository) CreateBlock(block *model.Block) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(block).Error
}

func (r *GormSocialRepository) DeleteBlock(blockerID, blockedID uuid.UUID) error {
	return r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&model.Block{}).Error
}

func (r *GormSocialRepository) CreateReport(report *model.Report) error {
	return r.db.Create(report).Error
}

func (r *GormSocialRepository) UserExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *GormSocialRepository) PostExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormSocialRepository) CommentExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
