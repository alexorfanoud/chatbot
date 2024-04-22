package model

import "time"

type Conversation struct {
	ID        int64
	SessionID int64
	Workflow  Workflow
	CreatedAt time.Time
}

type ConversationDTO struct {
	ID                  int64
	SessionID           int64
	WFExecutionContext  *WorkflowExecutionContext
	ConversationHistory []*Request
	CreatedAt           time.Time
}

func (c *ConversationDTO) AddRequestToHistory(r *Request) {
	if len(c.ConversationHistory) == 0 || c.ConversationHistory == nil {
		c.ConversationHistory = []*Request{}
	}

	c.ConversationHistory = append(c.ConversationHistory, r)
}
