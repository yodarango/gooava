package models

type Pantry struct {
	Id        uint   `json:id`
	UserId    string `json:user_id`
	CreatedAt string `json:created_at`
}
