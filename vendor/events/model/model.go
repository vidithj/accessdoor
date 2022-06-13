package model

type Events struct {
	Id       string             `json:"_id,omitempty"`
	Revision string             `json:"_rev,omitempty"`
	Username string             `json:"username"`
	Events   []map[string]int64 `json:"events"`
}

type UpdateEventRequest struct {
	Username string           `json:"username"`
	Event    map[string]int64 `json:"event"`
}
