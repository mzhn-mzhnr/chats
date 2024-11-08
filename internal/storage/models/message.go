package models

type NewMessage struct {
	ConversationId string
	IsUser         bool
	Body           string
}
