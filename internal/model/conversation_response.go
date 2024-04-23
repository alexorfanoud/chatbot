package model

type ConversationResponse struct {
	IsEOF                bool
	Content              string
	WorkflowFulfillments []WorkflowFulfillment
}

type WorkflowFulfillment struct {
	Name      string
	Arguments map[string]string
}
