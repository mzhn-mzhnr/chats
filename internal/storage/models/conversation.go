package models

import "time"

type Conversation struct {
	Id        string
	Name      *string
	CreatedAt time.Time
}
