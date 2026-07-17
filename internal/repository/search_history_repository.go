package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type SearchHistoryRepository interface {
	SaveSearch(userID uuid.UUID, query, queryType string) error
	ListHistory(userID uuid.UUID, limit int) ([]model.SearchHistory, error)
	DeleteHistoryItem(id, userID uuid.UUID) error
	DeleteAllHistory(userID uuid.UUID) error
}

type GormSearchHistoryRepository struct{ db *gorm.DB }

func NewSearchHistoryRepository(db *gorm.DB) SearchHistoryRepository {
	return &GormSearchHistoryRepository{db: db}
}

func (r *GormSearchHistoryRepository) SaveSearch(userID uuid.UUID, query, queryType string) error {
	// Upsert: update created_at if same user+query exists
	history := model.SearchHistory{
		UserID:    userID,
		Query:     query,
		QueryType: queryType,
	}
	return r.db.Where("user_id = ? AND query = ?", userID, query).
		Assign(model.SearchHistory{QueryType: queryType}).
		FirstOrCreate(&history).Error
}

func (r *GormSearchHistoryRepository) ListHistory(userID uuid.UUID, limit int) ([]model.SearchHistory, error) {
	var items []model.SearchHistory
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *GormSearchHistoryRepository) DeleteHistoryItem(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.SearchHistory{}).Error
}

func (r *GormSearchHistoryRepository) DeleteAllHistory(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}
