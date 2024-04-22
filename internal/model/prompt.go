package model

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

type Prompt struct {
	Workflow Workflow
	Text     string
	Tools    string
}

type PromptDTO struct {
	Workflow Workflow
	Text     string
	Tools    []openai.Tool
}

func (p *Prompt) ToPromptDTO() *PromptDTO {
	var tools []openai.Tool
	json.Unmarshal([]byte(p.Tools), &tools)

	return &PromptDTO{Workflow: p.Workflow, Text: p.Text, Tools: tools}

}
