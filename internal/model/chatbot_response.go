package model

type ChatbotResponse struct {
	IsEOF                bool
	Content              string
	WorkflowFulfillments []WorkflowFulfillment
}
