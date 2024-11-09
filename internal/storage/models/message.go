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

type Answer struct {
	Message
	AnswerMeta
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
