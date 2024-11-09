package models

import "time"

type Message struct {
	ConversationId string
	Body           string
	CreatedAt      time.Time
}

type Question struct {
	Message
}

type ChatHistoryEntry struct {
	IsUser bool
	Body   string
}

type RagRequest struct {
	Input       string `json:"input"`
	ChatHistory []ChatHistoryEntry
}

type RagResponse struct {
	Answer  string
	Sources []AnswerMeta
}

type Answer struct {
	Message
	Sources []AnswerMeta
}

type AnswerMetaSave struct {
	MessageId int
	AnswerMeta
}

type AnswerMeta struct {
	Filename string
	Slide    int
	FileId   string
}
