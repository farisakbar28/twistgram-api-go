package dto

import "time"

type SearchHistoryItem struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	QueryType string    `json:"query_type"`
	CreatedAt time.Time `json:"created_at"`
}

type DeleteSearchHistoryRequest struct {
	ID string `json:"id" binding:"omitempty,uuid"`
}
