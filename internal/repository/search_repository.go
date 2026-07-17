package repository

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type SearchRepository interface {
	SearchUsers(viewerID, query string, limit int) ([]model.User, error)
	SearchHashtags(viewerID, query string, limit int) ([]model.Hashtag, error)
	ListPostsByHashtag(tag string, viewerID uuid.UUID, page, limit int) ([]model.Post, int64, error)
}

type GormSearchRepository struct{ db *gorm.DB }

func NewSearchRepository(db *gorm.DB) SearchRepository { return &GormSearchRepository{db: db} }

func (r *GormSearchRepository) SearchUsers(viewerID, query string, limit int) ([]model.User, error) {
	q := strings.TrimSpace(query)
	db := r.db.Where("(LOWER(username) LIKE ? OR LOWER(name) LIKE ?)", "%"+strings.ToLower(q)+"%", "%"+strings.ToLower(q)+"%")
	if viewerID != "" {
		db = db.Where("id NOT IN (SELECT blocked_id FROM blocks WHERE blocker_id = ?) AND id NOT IN (SELECT blocker_id FROM blocks WHERE blocked_id = ?)", viewerID, viewerID)
	}
	var users []model.User
	err := db.Order("username ASC").Limit(limit).Find(&users).Error
	return users, err
}

func (r *GormSearchRepository) SearchHashtags(viewerID, query string, limit int) ([]model.Hashtag, error) {
	q := strings.TrimSpace(strings.TrimPrefix(query, "#"))
	var hashtags []model.Hashtag
	// Only count posts whose owner is public OR viewer is an accepted follower
	db := r.db.Table("hashtags").
		Select("hashtags.id, hashtags.tag, hashtags.created_at, COUNT(ph.post_id) AS post_count").
		Joins("LEFT JOIN post_hashtags ph ON ph.hashtag_id = hashtags.id").
		Joins("LEFT JOIN posts p ON p.id = ph.post_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN users u ON u.id = p.user_id")
	if viewerID != "" {
		db = db.Where(
			"LOWER(hashtags.tag) LIKE ? AND (p.id IS NULL OR u.is_private = false OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = ? AND f.following_id = p.user_id AND f.status = ?))",
			"%"+strings.ToLower(q)+"%", viewerID, "accepted",
		)
	} else {
		db = db.Where("LOWER(hashtags.tag) LIKE ? AND (p.id IS NULL OR u.is_private = false)", "%"+strings.ToLower(q)+"%")
	}
	err := db.
		Group("hashtags.id, hashtags.tag, hashtags.created_at").
		Order("post_count DESC, hashtags.tag ASC").
		Limit(limit).
		Scan(&hashtags).Error
	return hashtags, err
}

func (r *GormSearchRepository) ListPostsByHashtag(tag string, viewerID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	q := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
	var total int64

	query := r.db.Model(&model.Post{}).
		Joins("JOIN post_hashtags ph ON ph.post_id = posts.id").
		Joins("JOIN hashtags h ON h.id = ph.hashtag_id AND LOWER(h.tag) = ?", strings.ToLower(q)).
		Joins("JOIN users u ON u.id = posts.user_id").
		Where("posts.deleted_at IS NULL").
		Where("u.is_private = false OR u.id = ? OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = ? AND f.following_id = u.id AND f.status = ?)", viewerID, viewerID, "accepted")

	if viewerID != uuid.Nil {
		vid := viewerID.String()
		query = query.Where("u.id NOT IN (SELECT blocked_id FROM blocks WHERE blocker_id = ?) AND u.id NOT IN (SELECT blocker_id FROM blocks WHERE blocked_id = ?)", vid, vid)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []model.Post
	err := query.
		Preload("User").
		Preload("Media").
		Order("posts.created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&posts).Error

	return posts, total, err
}
