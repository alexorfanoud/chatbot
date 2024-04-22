package conversation

import (
	"chat/internal/api/middleware"
	"chat/internal/data/dao"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

type Chatbot interface {
	handleRequest(context.Context) error
}

type ChatbotImpl struct {
}

var (
	once             sync.Once
	client           *openai.Client
	toolExecutionMap = map[string]func(context.Context, *model.SessionDTO, map[string]string) string{
		"submit_review": func(ctx context.Context, s *model.SessionDTO, args map[string]string) string {
			return SubmitReview(ctx, s, args)
		},
		"complete_return": func(ctx context.Context, s *model.SessionDTO, args map[string]string) string {
			return CompleteReturn(ctx, s, args)
		},
	}
)

func getClient() *openai.Client {
	once.Do(func() { client = openai.NewClient(os.Getenv("OPENAI_TOKEN")) })
	return client
}

func PopulatePromptVariables(prompt string, variables map[string]string) string {
	for key := range variables {
		prompt = strings.ReplaceAll(prompt, "$"+key, variables[key])
	}
	return prompt
}

func (*ChatbotImpl) handleRequest(ctx context.Context, session *model.SessionDTO) (string, error) {
	activeConv := session.GetActiveConversation()
	workflowExecCtx := activeConv.WFExecutionContext
	prompt, err := dao.GetPromptByWorkflow(ctx, int(workflowExecCtx.Workflow))
	if err != nil {
		return "", err
	}
	prompt.Text = PopulatePromptVariables(prompt.Text, workflowExecCtx.PromptVariables)

	resp, err := gptRequest(ctx, getClient(), prompt, workflowExecCtx, session)
	if err != nil {
		return "", err
	}

	for _, toolCall := range resp.Choices[0].Message.ToolCalls {
		var args map[string]string
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		res := toolExecutionMap[toolCall.Function.Name](ctx, session, args)
		return res, nil
	}

	// TODO streaming?
	return resp.Choices[0].Message.Content, nil
}

func createRequest(prompt *model.PromptDTO, session *model.SessionDTO, workflowExecCtx *model.WorkflowExecutionContext, streaming bool) openai.ChatCompletionRequest {
	chatCompletionMessages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt.Text}}
	for iconv := max(0, len(session.Conversations)-1-workflowExecCtx.ContextWindow); iconv < len(session.Conversations); iconv++ {
		for _, req := range session.Conversations[iconv].ConversationHistory {
			if req.Question != "" {
				chatCompletionMessages = append(chatCompletionMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: req.Question})
			}

			if req.Answer != "" {
				chatCompletionMessages = append(chatCompletionMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: req.Answer})
			}
		}
	}

	return openai.ChatCompletionRequest{
		Model:    openai.GPT4Turbo,
		Tools:    prompt.Tools,
		Messages: chatCompletionMessages, Stream: streaming}

}

func gptRequest(ctx context.Context, client *openai.Client, prompt *model.PromptDTO, workflowExecCtx *model.WorkflowExecutionContext, session *model.SessionDTO) (*openai.ChatCompletionResponse, error) {
	// utils.Log(ctx, fmt.Sprintf("ChatCompletion request: %+v\n", createRequest(prompt, session, workflowExecCtx, false)))
	ctx, span := middleware.Tracer.Start(ctx, "gptRequest")
	defer span.End()
	resp, err := client.CreateChatCompletion(ctx, createRequest(prompt, session, workflowExecCtx, false))
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("ChatCompletion error: %v\n", err))
		return nil, err
	}

	// utils.Log(ctx, fmt.Sprintf("ChatCompletion response: %+v\n", resp))
	return &resp, nil
}
