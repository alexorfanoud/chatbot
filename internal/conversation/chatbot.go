package conversation

import (
	"chat/internal/api/middleware"
	"chat/internal/data/dao"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

type Chatbot interface {
	handleRequest(context.Context, *model.SessionDTO) (model.ChatbotResponse, error)
	handleRequestStreaming(context.Context, *model.SessionDTO) (<-chan model.ChatbotResponse, error)
}

var (
	once           sync.Once
	openAIClient   *openai.Client
	oneshotChatbot = ChatbotImpl{}
)

func getOpenAIClient() *openai.Client {
	once.Do(func() { openAIClient = openai.NewClient(os.Getenv("OPENAI_TOKEN")) })
	return openAIClient
}

func populatePromptVariables(prompt string, variables map[string]string) string {
	for key := range variables {
		prompt = strings.ReplaceAll(prompt, "$"+key, variables[key])
	}
	return prompt
}

type ChatbotImpl struct{}

func (*ChatbotImpl) handleRequest(ctx context.Context, session *model.SessionDTO) (*model.ChatbotResponse, error) {
	activeConv := session.GetActiveConversation()
	workflowExecCtx := activeConv.WFExecutionContext
	prompt, err := dao.GetPromptByWorkflow(ctx, int(workflowExecCtx.Workflow))
	if err != nil {
		return nil, err
	}
	prompt.Text = populatePromptVariables(prompt.Text, workflowExecCtx.PromptVariables)

	resp, err := gptRequest(ctx, getOpenAIClient(), prompt, workflowExecCtx, session)
	if err != nil {
		return nil, err
	}

	chatbotRes := model.ChatbotResponse{
		IsEOF:                true,
		Content:              resp.Choices[0].Message.Content,
		WorkflowFulfillments: make([]model.WorkflowFulfillment, 0),
	}

	for _, toolCall := range resp.Choices[0].Message.ToolCalls {
		var args map[string]string
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		chatbotRes.WorkflowFulfillments = append(chatbotRes.WorkflowFulfillments, model.WorkflowFulfillment{Name: toolCall.Function.Name, Arguments: args})
	}

	return &chatbotRes, nil
}

func (*ChatbotImpl) handleRequestStreaming(ctx context.Context, session *model.SessionDTO) (<-chan *model.ChatbotResponse, error) {
	activeConv := session.GetActiveConversation()
	workflowExecCtx := activeConv.WFExecutionContext
	prompt, err := dao.GetPromptByWorkflow(ctx, int(workflowExecCtx.Workflow))
	if err != nil {
		return nil, err
	}
	prompt.Text = populatePromptVariables(prompt.Text, workflowExecCtx.PromptVariables)

	stream, err := gptRequestStream(ctx, getOpenAIClient(), prompt, workflowExecCtx, session)
	if err != nil {
		return nil, err
	}

	resCh := make(chan *model.ChatbotResponse)
	go func() {
		defer close(resCh)

		toolCalls := make([]openai.FunctionCall, 0)
		totalResp := ""
		for {
			res, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				utils.Log(ctx, fmt.Sprintf("Error getting gpt stream response: %s", err.Error()))
				return
			}
			content := res.Choices[0].Delta.Content
			chatbotRes := model.ChatbotResponse{
				IsEOF:                false,
				Content:              content,
				WorkflowFulfillments: make([]model.WorkflowFulfillment, 0),
			}
			totalResp += content
			resCh <- &chatbotRes

			// Gather delta tool calls
			processStreamToolCalls(&toolCalls, res)
		}

		chatbotResFinal := model.ChatbotResponse{
			IsEOF:                true,
			Content:              totalResp,
			WorkflowFulfillments: make([]model.WorkflowFulfillment, 0),
		}
		// If tool call was triggered, send over the response
		var args map[string]string
		for _, toolCall := range toolCalls {
			json.Unmarshal([]byte(toolCall.Arguments), &args)
			chatbotResFinal.WorkflowFulfillments = append(chatbotResFinal.WorkflowFulfillments, model.WorkflowFulfillment{Name: toolCall.Name, Arguments: args})
		}

		resCh <- &chatbotResFinal
	}()

	return resCh, nil
}

func processStreamToolCalls(tc *[]openai.FunctionCall, streamResp openai.ChatCompletionStreamResponse) {
	for idx, toolCall := range streamResp.Choices[0].Delta.ToolCalls {
		if idx == len(*tc) {
			*tc = append(*tc, openai.FunctionCall{})
		}
		if toolCall.Function.Name != "" {
			(*tc)[idx].Name = toolCall.Function.Name
		}
		if toolCall.Function.Arguments != "" {
			(*tc)[idx].Arguments += toolCall.Function.Arguments
		}
	}
}

func createOpenAIRequest(prompt *model.PromptDTO, session *model.SessionDTO, workflowExecCtx *model.WorkflowExecutionContext, streaming bool) openai.ChatCompletionRequest {
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
	resp, err := client.CreateChatCompletion(ctx, createOpenAIRequest(prompt, session, workflowExecCtx, false))
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("ChatCompletion error: %v\n", err))
		return nil, err
	}

	// utils.Log(ctx, fmt.Sprintf("ChatCompletion response: %+v\n", resp))
	return &resp, nil
}

func gptRequestStream(ctx context.Context, client *openai.Client, prompt *model.PromptDTO, workflowExecCtx *model.WorkflowExecutionContext, session *model.SessionDTO) (*openai.ChatCompletionStream, error) {
	// utils.Log(ctx, fmt.Sprintf("ChatCompletion request: %+v\n", createOpenAIRequest(prompt, session, workflowExecCtx, false)))
	ctx, span := middleware.Tracer.Start(ctx, "gptRequest")
	defer span.End()
	stream, err := client.CreateChatCompletionStream(ctx, createOpenAIRequest(prompt, session, workflowExecCtx, true))
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("ChatCompletion error: %v\n", err))
		return nil, err
	}

	return stream, nil
}
