package domain

import "time"

type NewMessage struct {
	Body           string
	UserId         *string
	ConversationId *string
}

type SentMessage struct {
	ConversationId string
}

type Message struct {
	Id             int
	Body           string
	IsUser         bool
	ConversationId string
	CreatedAt      time.Time
}

type MessageType struct {
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type HandledMessage struct {
	ConversationId *string `json:"conversation_id,omitempty"`
	UserId         string  `json:"user_id"`

	Question MessageType `json:"question"`
	Answer   MessageType `json:"answer"`
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
