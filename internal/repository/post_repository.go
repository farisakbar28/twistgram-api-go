package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type PostRepository interface {
	CreatePostWithMedia(post *model.Post, media []model.PostMedia) error
	GetPostByID(id uuid.UUID) (*model.Post, error)

	GetPostWithMedia(id uuid.UUID) (*model.Post, error)
	UpdatePost(post *model.Post) error
	ListFeed(userID uuid.UUID, page, limit int) ([]model.Post, int64, error)
	ListUserPosts(userID uuid.UUID, page, limit int) ([]model.Post, int64, error)
	DeletePost(id uuid.UUID, userID uuid.UUID) error
	SetArchived(id uuid.UUID, userID uuid.UUID, archived bool) error
	PostExists(id uuid.UUID) (bool, error)
	OwnsPost(id, userID uuid.UUID) (bool, error)
	UserHashtagUpsert(tags []string) ([]model.Hashtag, error)
	ReplacePostHashtags(postID uuid.UUID, hashtags []model.Hashtag) error
	ReplacePostTags(postID uuid.UUID, taggedUserIDs []uuid.UUID) error
}

type GormPostRepository struct{ db *gorm.DB }

func NewPostRepository(db *gorm.DB) PostRepository { return &GormPostRepository{db: db} }

func (r *GormPostRepository) CreatePostWithMedia(post *model.Post, media []model.PostMedia) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil { return err }
		for i := range media { media[i].PostID = post.ID }
		if len(media) > 0 {
			if err := tx.Create(&media).Error; err != nil { return err }
		}
		return nil
	})
}

func (r *GormPostRepository) GetPostByID(id uuid.UUID) (*model.Post, error) {
	var post model.Post
	if err := r.db.First(&post, "id = ? AND deleted_at IS NULL", id).Error; err != nil { return nil, err }
	return &post, nil
}

func (r *GormPostRepository) GetPostWithMedia(id uuid.UUID) (*model.Post, error) {
	var post model.Post
	if err := r.db.Preload("Media").Preload("Tags").Preload("Hashtags").First(&post, "id = ? AND deleted_at IS NULL", id).Error; err != nil { return nil, err }
	return &post, nil
}

func (r *GormPostRepository) UpdatePost(post *model.Post) error { return r.db.Save(post).Error }

func (r *GormPostRepository) ListFeed(userID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	var total int64
	query := r.db.Model(&model.Post{}).
		Joins("JOIN follows ON follows.following_id = posts.user_id").
		Where("follows.follower_id = ? AND follows.status = ? AND posts.deleted_at IS NULL AND posts.is_archived = false", userID, "accepted")
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var posts []model.Post
	err := query.Preload("Media").Order("posts.created_at DESC").Offset((page-1)*limit).Limit(limit).Find(&posts).Error
	return posts, total, err
}

func (r *GormPostRepository) ListUserPosts(userID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	var total int64
	query := r.db.Model(&model.Post{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var posts []model.Post
	err := query.Preload("Media").Order("created_at DESC").Offset((page-1)*limit).Limit(limit).Find(&posts).Error
	return posts, total, err
}

func (r *GormPostRepository) DeletePost(id uuid.UUID, userID uuid.UUID) error {
	res := r.db.Model(&model.Post{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Update("deleted_at", gorm.Expr("now()"))
	if res.Error != nil { return res.Error }
	if res.RowsAffected == 0 { return gorm.ErrRecordNotFound }
	return nil
}

func (r *GormPostRepository) SetArchived(id uuid.UUID, userID uuid.UUID, archived bool) error {
	res := r.db.Model(&model.Post{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Update("is_archived", archived)
	if res.Error != nil { return res.Error }
	if res.RowsAffected == 0 { return gorm.ErrRecordNotFound }
	return nil
}

func (r *GormPostRepository) PostExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormPostRepository) OwnsPost(id, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Count(&count).Error
	return count > 0, err
}

func (r *GormPostRepository) UserHashtagUpsert(tags []string) ([]model.Hashtag, error) {
	unique := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(tag, "#")))
		if tag != "" { unique[tag] = struct{}{} }
	}
	out := make([]model.Hashtag, 0, len(unique))
	for tag := range unique {
		h := model.Hashtag{Tag: tag}
		if err := r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tag"}}, DoNothing: true}).Create(&h).Error; err != nil { return nil, err }
		if err := r.db.First(&h, "tag = ?", tag).Error; err != nil { return nil, err }
		out = append(out, h)
	}
	return out, nil
}

func (r *GormPostRepository) ReplacePostHashtags(postID uuid.UUID, hashtags []model.Hashtag) error { return nil }
func (r *GormPostRepository) ReplacePostTags(postID uuid.UUID, taggedUserIDs []uuid.UUID) error { return nil }

var _ = errors.New
