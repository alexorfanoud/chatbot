package model

type WorkflowExecutionContext struct {
	Workflow        Workflow
	PromptVariables map[string]string
	ContextWindow   int
}
