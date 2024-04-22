package model

import "time"

type Request struct {
	ID             int64
	ConversationID int64
	Question       string
	Answer         string
	CreatedAt      time.Time
}
