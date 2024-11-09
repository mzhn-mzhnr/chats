package domain

import (
	"time"
)

type NewMessageRequest struct {
	Body           string
	ConversationId string
	CreatedAt      time.Time
}

type StreamMessageRequest struct {
	NewMessageRequest
	EventCh chan<- []byte
}

type SentMessage struct {
	Answer  string       `json:"answer"`
	Sources []AnswerMeta `json:"sources"`
}

type Message struct {
	Id             int
	Body           string
	IsUser         bool
	ConversationId string
	CreatedAt      time.Time
	Meta           *AnswerMeta
}

type AnswerMeta struct {
	FileId   string `json:"fileId"`
	FileName string `json:"fileName"`
	Slidenum int    `json:"slideNum"`
}

type MessageType struct {
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type HandledMessage struct {
	ConversationId string       `json:"conversation_id"`
	Question       *MessageType `json:"question"`
	Answer         *MessageType `json:"answer"`
}

func (m *HandledMessage) Valid() bool {
	return !(m.ConversationId == "" || m.Question == nil || m.Answer == nil)
}

type NewConversation struct {
	UserId string
}

type Conversation struct {
	Id        string
	Name      *string
	CreatedAt time.Time
}

type ConversationContent struct {
	Conversation
	Messages []*Message
}

type ConversationsFilter struct {
	UserId string
}
