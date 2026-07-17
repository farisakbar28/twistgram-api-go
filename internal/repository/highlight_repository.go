package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type HighlightRepository interface {
	CreateHighlight(highlight *model.Highlight) error
	GetHighlightByID(id uuid.UUID) (*model.Highlight, error)
	ListHighlightsByUser(userID uuid.UUID) ([]model.Highlight, error)
	UpdateHighlight(highlight *model.Highlight) error
	DeleteHighlight(id uuid.UUID) error
	AddStoryToHighlight(highlightID, storyID uuid.UUID) error
	RemoveStoryFromHighlight(highlightID, storyID uuid.UUID) error
	ListHighlightStories(highlightID uuid.UUID) ([]model.Story, error)
	HighlightExists(id uuid.UUID) (bool, error)
	OwnsHighlight(highlightID, userID uuid.UUID) (bool, error)
}

type GormHighlightRepository struct{ db *gorm.DB }

func NewHighlightRepository(db *gorm.DB) HighlightRepository {
	return &GormHighlightRepository{db: db}
}

func (r *GormHighlightRepository) CreateHighlight(highlight *model.Highlight) error {
	return r.db.Create(highlight).Error
}

func (r *GormHighlightRepository) GetHighlightByID(id uuid.UUID) (*model.Highlight, error) {
	var h model.Highlight
	if err := r.db.First(&h, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *GormHighlightRepository) ListHighlightsByUser(userID uuid.UUID) ([]model.Highlight, error) {
	var highlights []model.Highlight
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&highlights).Error
	return highlights, err
}

func (r *GormHighlightRepository) UpdateHighlight(highlight *model.Highlight) error {
	return r.db.Save(highlight).Error
}

func (r *GormHighlightRepository) DeleteHighlight(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("highlight_id = ?", id).Delete(&model.HighlightStory{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Highlight{}, "id = ?", id).Error
	})
}

func (r *GormHighlightRepository) AddStoryToHighlight(highlightID, storyID uuid.UUID) error {
	hs := model.HighlightStory{HighlightID: highlightID, StoryID: storyID}
	return r.db.Create(&hs).Error
}

func (r *GormHighlightRepository) RemoveStoryFromHighlight(highlightID, storyID uuid.UUID) error {
	return r.db.Where("highlight_id = ? AND story_id = ?", highlightID, storyID).Delete(&model.HighlightStory{}).Error
}

func (r *GormHighlightRepository) ListHighlightStories(highlightID uuid.UUID) ([]model.Story, error) {
	var stories []model.Story
	err := r.db.
		Joins("JOIN highlight_stories hs ON hs.story_id = stories.id").
		Where("hs.highlight_id = ?", highlightID).
		Order("stories.created_at DESC").
		Find(&stories).Error
	return stories, err
}

func (r *GormHighlightRepository) HighlightExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Highlight{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *GormHighlightRepository) OwnsHighlight(highlightID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Highlight{}).Where("id = ? AND user_id = ?", highlightID, userID).Count(&count).Error
	return count > 0, err
}
