package repository

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type PostRepository interface {
	CreatePostWithMediaAndTags(post *model.Post, media []model.PostMedia, tags []model.PostTag, hashtags []model.Hashtag) error
	UpdateCaption(id uuid.UUID, userID uuid.UUID, caption *string) error
	GetPostByID(id uuid.UUID) (*model.Post, error)

	GetPostWithMedia(id uuid.UUID) (*model.Post, error)
	UpdatePost(post *model.Post) error
	ListFeed(userID uuid.UUID, page, limit int) ([]model.Post, int64, error)
	ListUserPosts(userID uuid.UUID, page, limit int) ([]model.Post, int64, error)
	ListGlobalFeed(limit int) ([]model.Post, int64, error)
	DeletePost(id uuid.UUID, userID uuid.UUID) error
	SetArchived(id uuid.UUID, userID uuid.UUID, archived bool) error
	PostExists(id uuid.UUID) (bool, error)
	OwnsPost(id, userID uuid.UUID) (bool, error)
	UserHashtagUpsert(tags []string) ([]model.Hashtag, error)
	CreateNotification(notification *model.Notification) error
	DeletePostTag(postID, taggedUserID uuid.UUID) error
	// Visibility helpers for user profile posts
	IsUserPrivate(userID uuid.UUID) (bool, error)
	IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error)
	IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error)
}

type GormPostRepository struct{ db *gorm.DB }

func NewPostRepository(db *gorm.DB) PostRepository { return &GormPostRepository{db: db} }

func (r *GormPostRepository) CreatePostWithMediaAndTags(post *model.Post, media []model.PostMedia, tags []model.PostTag, hashtags []model.Hashtag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		if len(media) > 0 {
			for i := range media {
				media[i].PostID = post.ID
			}
			if err := tx.Create(&media).Error; err != nil {
				return err
			}
		}

		if len(tags) > 0 {
			for i := range tags {
				tags[i].PostID = post.ID
			}
			if err := tx.Create(&tags).Error; err != nil {
				return err
			}
		}

		if len(hashtags) > 0 {
			postHashtags := make([]model.PostHashtag, 0, len(hashtags))
			for _, h := range hashtags {
				var existing model.Hashtag
				if err := tx.Where("tag = ?", h.Tag).First(&existing).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						existing = model.Hashtag{Tag: h.Tag}
						if err := tx.Create(&existing).Error; err != nil {
							return err
						}
					} else {
						return err
					}
				}
				postHashtags = append(postHashtags, model.PostHashtag{PostID: post.ID, HashtagID: existing.ID})
			}
			if err := tx.Create(&postHashtags).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormPostRepository) UpdateCaption(id uuid.UUID, userID uuid.UUID, caption *string) error {
	res := r.db.Model(&model.Post{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Update("caption", caption)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormPostRepository) GetPostByID(id uuid.UUID) (*model.Post, error) {
	var post model.Post
	if err := r.db.First(&post, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *GormPostRepository) GetPostWithMedia(id uuid.UUID) (*model.Post, error) {
	var post model.Post
	if err := r.db.Preload("Media").Preload("Tags").Preload("Hashtags").First(&post, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *GormPostRepository) UpdatePost(post *model.Post) error { return r.db.Save(post).Error }

func (r *GormPostRepository) ListFeed(userID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	var total int64
	query := r.db.Model(&model.Post{}).
		Joins("JOIN follows ON follows.following_id = posts.user_id").
		Where("follows.follower_id = ? AND follows.status = ? AND posts.deleted_at IS NULL AND posts.is_archived = false", userID, "accepted")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var posts []model.Post
	if total == 0 {
		return posts, 0, nil
	} // Fallback handled by service
	err := query.Preload("User").Preload("Media").Preload("Tags").Preload("Hashtags").Order("posts.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&posts).Error
	return posts, total, err
}

func (r *GormPostRepository) ListGlobalFeed(limit int) ([]model.Post, int64, error) {
	var total int64
	query := r.db.Model(&model.Post{}).
		Joins("JOIN users u ON u.id = posts.user_id").
		Where("posts.deleted_at IS NULL AND posts.is_archived = false AND u.is_private = false")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var posts []model.Post
	err := query.Preload("User").Preload("Media").Order("posts.created_at DESC").Limit(limit).Find(&posts).Error
	return posts, total, err
}

func (r *GormPostRepository) ListUserPosts(userID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	var total int64
	query := r.db.Model(&model.Post{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var posts []model.Post
	err := query.Preload("Media").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&posts).Error
	return posts, total, err
}

func (r *GormPostRepository) DeletePost(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verify ownership first
		var count int64
		if err := tx.Model(&model.Post{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// Soft-delete related records
		_ = tx.Where("post_id = ?", id).Delete(&model.Like{}).Error
		_ = tx.Where("post_id = ?", id).Delete(&model.Comment{}).Error
		_ = tx.Where("post_id = ?", id).Delete(&model.SavedPost{}).Error
		_ = tx.Where("post_id = ?", id).Delete(&model.PostTag{}).Error
		_ = tx.Where("post_id = ?", id).Delete(&model.PostHashtag{}).Error

		// Soft-delete the post itself
		return tx.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", gorm.Expr("now()")).Error
	})
}

func (r *GormPostRepository) SetArchived(id uuid.UUID, userID uuid.UUID, archived bool) error {
	res := r.db.Model(&model.Post{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Update("is_archived", archived)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
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
		if tag != "" {
			unique[tag] = struct{}{}
		}
	}
	out := make([]model.Hashtag, 0, len(unique))
	for tag := range unique {
		h := model.Hashtag{Tag: tag}
		if err := r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tag"}}, DoNothing: true}).Create(&h).Error; err != nil {
			return nil, err
		}
		if err := r.db.First(&h, "tag = ?", tag).Error; err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

func (r *GormPostRepository) DeletePostTag(postID, taggedUserID uuid.UUID) error {
	return r.db.Where("post_id = ? AND tagged_user_id = ?", postID, taggedUserID).Delete(&model.PostTag{}).Error
}

func (r *GormPostRepository) CreateNotification(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

func (r *GormPostRepository) IsUserPrivate(userID uuid.UUID) (bool, error) {
	var isPrivate bool
	err := r.db.Model(&model.User{}).Select("is_private").Where("id = ?", userID).Scan(&isPrivate).Error
	return isPrivate, err
}

func (r *GormPostRepository) IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Block{}).Where(
		"(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
		userA, userB, userB, userA,
	).Count(&count).Error
	return count > 0, err
}

func (r *GormPostRepository) IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").
		Count(&count).Error
	return count > 0, err
}
