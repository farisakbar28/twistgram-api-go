package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type SearchHistoryRepository interface {
	SaveSearch(ctx context.Context, userID uuid.UUID, query, queryType string) error
	ListHistory(ctx context.Context, userID uuid.UUID, limit int) ([]model.SearchHistory, error)
	DeleteHistoryItem(ctx context.Context, id, userID uuid.UUID) error
	DeleteAllHistory(ctx context.Context, userID uuid.UUID) error
}

type GormSearchHistoryRepository struct{ db *gorm.DB }

func NewSearchHistoryRepository(db *gorm.DB) SearchHistoryRepository {
	return &GormSearchHistoryRepository{db: db}
}

func (r *GormSearchHistoryRepository) SaveSearch(ctx context.Context, userID uuid.UUID, query, queryType string) error {
	history := model.SearchHistory{
		UserID:    userID,
		Query:     query,
		QueryType: queryType,
	}
	return r.db.WithContext(ctx).Where("user_id = ? AND query = ?", userID, query).
		Assign(model.SearchHistory{QueryType: queryType}).
		FirstOrCreate(&history).Error
}

func (r *GormSearchHistoryRepository) ListHistory(ctx context.Context, userID uuid.UUID, limit int) ([]model.SearchHistory, error) {
	var items []model.SearchHistory
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *GormSearchHistoryRepository) DeleteHistoryItem(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.SearchHistory{}).Error
}

func (r *GormSearchHistoryRepository) DeleteAllHistory(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}
