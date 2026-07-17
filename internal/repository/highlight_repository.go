package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type HighlightRepository interface {
	CreateHighlight(ctx context.Context, highlight *model.Highlight) error
	GetHighlightByID(ctx context.Context, id uuid.UUID) (*model.Highlight, error)
	ListHighlightsByUser(ctx context.Context, userID uuid.UUID) ([]model.Highlight, error)
	UpdateHighlight(ctx context.Context, highlight *model.Highlight) error
	DeleteHighlight(ctx context.Context, id uuid.UUID) error
	AddStoryToHighlight(ctx context.Context, highlightID, storyID uuid.UUID) error
	RemoveStoryFromHighlight(ctx context.Context, highlightID, storyID uuid.UUID) error
	ListHighlightStories(ctx context.Context, highlightID uuid.UUID) ([]model.Story, error)
	HighlightExists(ctx context.Context, id uuid.UUID) (bool, error)
	OwnsHighlight(ctx context.Context, highlightID, userID uuid.UUID) (bool, error)
}

type GormHighlightRepository struct{ db *gorm.DB }

func NewHighlightRepository(db *gorm.DB) HighlightRepository {
	return &GormHighlightRepository{db: db}
}

func (r *GormHighlightRepository) CreateHighlight(ctx context.Context, highlight *model.Highlight) error {
	return r.db.WithContext(ctx).Create(highlight).Error
}

func (r *GormHighlightRepository) GetHighlightByID(ctx context.Context, id uuid.UUID) (*model.Highlight, error) {
	var h model.Highlight
	if err := r.db.WithContext(ctx).First(&h, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *GormHighlightRepository) ListHighlightsByUser(ctx context.Context, userID uuid.UUID) ([]model.Highlight, error) {
	var highlights []model.Highlight
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&highlights).Error
	return highlights, err
}

func (r *GormHighlightRepository) UpdateHighlight(ctx context.Context, highlight *model.Highlight) error {
	return r.db.WithContext(ctx).Save(highlight).Error
}

func (r *GormHighlightRepository) DeleteHighlight(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("highlight_id = ?", id).Delete(&model.HighlightStory{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Highlight{}, "id = ?", id).Error
	})
}

func (r *GormHighlightRepository) AddStoryToHighlight(ctx context.Context, highlightID, storyID uuid.UUID) error {
	hs := model.HighlightStory{HighlightID: highlightID, StoryID: storyID}
	return r.db.WithContext(ctx).Create(&hs).Error
}

func (r *GormHighlightRepository) RemoveStoryFromHighlight(ctx context.Context, highlightID, storyID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("highlight_id = ? AND story_id = ?", highlightID, storyID).Delete(&model.HighlightStory{}).Error
}

func (r *GormHighlightRepository) ListHighlightStories(ctx context.Context, highlightID uuid.UUID) ([]model.Story, error) {
	var stories []model.Story
	err := r.db.WithContext(ctx).
		Joins("JOIN highlight_stories hs ON hs.story_id = stories.id").
		Where("hs.highlight_id = ?", highlightID).
		Order("stories.created_at DESC").
		Find(&stories).Error
	return stories, err
}

func (r *GormHighlightRepository) HighlightExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Highlight{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *GormHighlightRepository) OwnsHighlight(ctx context.Context, highlightID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Highlight{}).Where("id = ? AND user_id = ?", highlightID, userID).Count(&count).Error
	return count > 0, err
}
