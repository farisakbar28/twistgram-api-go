package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type StoryRepository interface {
	CreateStoryWithTags(ctx context.Context, story *model.Story, tags []model.StoryTag) error
	GetStoryByID(ctx context.Context, id uuid.UUID) (*model.Story, error)
	ListActiveFeedStories(ctx context.Context, userID uuid.UUID) ([]model.Story, error)
	FindActiveStoryByUserID(ctx context.Context, userID uuid.UUID) (*model.Story, error)
	DeleteExpiredStories(ctx context.Context) error
	RecordView(ctx context.Context, view *model.StoryView) error
	ListViewers(ctx context.Context, storyID uuid.UUID) ([]model.StoryView, error)
	IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	DeleteStory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetStoryOwner(ctx context.Context, storyID uuid.UUID) (uuid.UUID, error)
	CreateNotification(ctx context.Context, notification *model.Notification) error
}

type GormStoryRepository struct{ db *gorm.DB }

func NewStoryRepository(db *gorm.DB) StoryRepository { return &GormStoryRepository{db: db} }

func (r *GormStoryRepository) CreateStoryWithTags(ctx context.Context, story *model.Story, tags []model.StoryTag) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(story).Error; err != nil {
			return err
		}
		if len(tags) > 0 {
			for i := range tags {
				tags[i].StoryID = story.ID
			}
			if err := tx.Create(&tags).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormStoryRepository) GetStoryByID(ctx context.Context, id uuid.UUID) (*model.Story, error) {
	var s model.Story
	if err := r.db.WithContext(ctx).First(&s, "id = ? AND expires_at > ?", id, time.Now()).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormStoryRepository) ListActiveFeedStories(ctx context.Context, userID uuid.UUID) ([]model.Story, error) {
	var stories []model.Story
	err := r.db.WithContext(ctx).
		Where("stories.expires_at > ?", time.Now()).
		Where("stories.user_id = ? OR stories.user_id IN (SELECT following_id FROM follows WHERE follower_id = ? AND status = ?)", userID, userID, "accepted").
		Preload("User").
		Order("stories.created_at DESC").
		Find(&stories).Error
	return stories, err
}

func (r *GormStoryRepository) DeleteStory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Story{}).Where("id = ? AND user_id = ?", id, userID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Where("story_id = ?", id).Delete(&model.StoryView{}).Error; err != nil {
			return err
		}
		if err := tx.Where("story_id = ?", id).Delete(&model.StoryTag{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Message{}).Where("reply_to_story_id = ?", id).Update("reply_to_story_id", nil).Error; err != nil {
			return err
		}

		return tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Story{}).Error
	})
}

func (r *GormStoryRepository) FindActiveStoryByUserID(ctx context.Context, userID uuid.UUID) (*model.Story, error) {
	var s model.Story
	err := r.db.WithContext(ctx).Where("user_id = ? AND expires_at > ?", userID, time.Now()).First(&s).Error
	return &s, err
}

func (r *GormStoryRepository) DeleteExpiredStories(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&model.Story{}).Error
}

func (r *GormStoryRepository) RecordView(ctx context.Context, view *model.StoryView) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "story_id"}, {Name: "viewer_id"}},
		DoNothing: true,
	}).Create(view).Error
}

func (r *GormStoryRepository) ListViewers(ctx context.Context, storyID uuid.UUID) ([]model.StoryView, error) {
	var views []model.StoryView
	err := r.db.WithContext(ctx).Where("story_id = ?", storyID).Preload("Viewer").Order("viewed_at DESC").Find(&views).Error
	return views, err
}

func (r *GormStoryRepository) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Block{}).Where(
		"(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
		userA, userB, userB, userA,
	).Count(&count).Error
	return count > 0, err
}

func (r *GormStoryRepository) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").Count(&count).Error
	return count > 0, err
}

func (r *GormStoryRepository) GetStoryOwner(ctx context.Context, storyID uuid.UUID) (uuid.UUID, error) {
	var s model.Story
	if err := r.db.WithContext(ctx).Select("user_id").First(&s, "id = ?", storyID).Error; err != nil {
		return uuid.Nil, err
	}
	return s.UserID, nil
}

func (r *GormStoryRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	return CreateNotificationHelper(ctx, r.db, notification)
}
