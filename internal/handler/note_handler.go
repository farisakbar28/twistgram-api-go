package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type NoteHandler struct{ service *service.NoteService }

func NewNoteHandlerWithService(svc *service.NoteService) *NoteHandler {
	return &NoteHandler{service: svc}
}

func (h *NoteHandler) Create(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	var req dto.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.service.CreateNote(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			response.BadRequest(c, "Invalid note content")
		} else {
			response.InternalError(c, "Failed to create note")
		}
		return
	}
	response.Created(c, gin.H{"note": res})
}

func (h *NoteHandler) GetActive(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	notes, err := h.service.GetActiveNotes(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to fetch notes")
		return
	}
	response.Success(c, gin.H{"notes": notes})
}

func (h *NoteHandler) Delete(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid note id")
		return
	}
	if err := h.service.DeleteNote(c.Request.Context(), userID, noteID); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not own this note")
		} else {
			response.InternalError(c, "Failed to delete note")
		}
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
