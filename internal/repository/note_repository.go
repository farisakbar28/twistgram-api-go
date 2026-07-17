package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type NoteRepository interface {
	CreateNote(ctx context.Context, note *model.Note) error
	GetActiveNotes(ctx context.Context, userID uuid.UUID) ([]model.Note, error)
	DeleteNote(ctx context.Context, noteID, userID uuid.UUID) error
	OwnsNote(ctx context.Context, noteID, userID uuid.UUID) (bool, error)
}

type GormNoteRepository struct{ db *gorm.DB }

func NewNoteRepository(db *gorm.DB) NoteRepository { return &GormNoteRepository{db: db} }

func (r *GormNoteRepository) CreateNote(ctx context.Context, note *model.Note) error {
	return r.db.WithContext(ctx).Create(note).Error
}

func (r *GormNoteRepository) GetActiveNotes(ctx context.Context, userID uuid.UUID) ([]model.Note, error) {
	var notes []model.Note
	// NTF-02 logic: visible to mutual followers (and self)
	// Phase 2 notes rule: mutual followers + active notes
	query := r.db.WithContext(ctx).Model(&model.Note{}).
		Where("notes.expires_at > ?", time.Now()).
		Where("notes.user_id = ? OR (EXISTS (SELECT 1 FROM follows f1 WHERE f1.follower_id = ? AND f1.following_id = notes.user_id AND f1.status = 'accepted') AND EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = notes.user_id AND f2.following_id = ? AND f2.status = 'accepted'))", userID, userID, userID).
		Preload("User").
		Order("notes.created_at DESC").
		Find(&notes)
	return notes, query.Error
}

func (r *GormNoteRepository) DeleteNote(ctx context.Context, noteID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", noteID, userID).Delete(&model.Note{}).Error
}

func (r *GormNoteRepository) OwnsNote(ctx context.Context, noteID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Note{}).Where("id = ? AND user_id = ?", noteID, userID).Count(&count).Error
	return count > 0, err
}
