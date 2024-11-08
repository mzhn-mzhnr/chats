package storage

import "errors"

var (
	ErrConversationsNotFound = errors.New("conversations not found")
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrMessagesNotFound      = errors.New("messages not found")
	ErrMessageNotFound       = errors.New("message not found")

	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadToken     = errors.New("bad token")
)
