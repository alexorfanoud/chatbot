package model

import (
	"time"
)

type Session struct {
	ID        int64
	UserID    int64
	CreatedAt time.Time
}

type SessionDTO struct {
	ID                    int64
	UserID                int64
	CreatedAt             time.Time
	ActiveConversationIdx int
	Conversations         []*ConversationDTO
	ChannelType           ChannelType
}

func (s *SessionDTO) GetActiveConversation() *ConversationDTO {
	if len(s.Conversations) == 0 || s.ActiveConversationIdx == -1 {
		return nil
	}
	return s.Conversations[s.ActiveConversationIdx]
}
