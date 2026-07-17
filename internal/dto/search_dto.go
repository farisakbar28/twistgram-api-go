package dto

type SearchUserItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type SearchHashtagItem struct {
	ID    string `json:"id"`
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

type SearchResponse struct {
	Users    []SearchUserItem    `json:"users"`
	Hashtags []SearchHashtagItem `json:"hashtags"`
}
