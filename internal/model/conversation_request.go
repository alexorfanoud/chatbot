package model

type ConversationRequest struct {
	Request                  string
	UserID                   int64
	ChannelType              ChannelType
	WorkflowExecutionContext *WorkflowExecutionContext
	NewConversation          bool
}
