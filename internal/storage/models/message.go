package models

import "time"

type NewMessage struct {
	ConversationId string
	IsUser         bool
	Body           string
	CreatedAt      time.Time
}
