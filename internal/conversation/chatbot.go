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
	handleRequest(context.Context, *model.SessionDTO) (<-chan string, error)
}

type ChatbotImpl struct {
}

type StreamingChatbotImpl struct {
}

var (
	once             sync.Once
	openAIClient     *openai.Client
	toolExecutionMap = map[string]func(context.Context, *model.SessionDTO, map[string]string) string{
		"submit_review": func(ctx context.Context, s *model.SessionDTO, args map[string]string) string {
			return SubmitProductReview(ctx, s, args)
		},
		"complete_return": func(ctx context.Context, s *model.SessionDTO, args map[string]string) string {
			return CompleteProductReturn(ctx, s, args)
		},
	}
	oneshotChatbot   = ChatbotImpl{}
	streamingChatbot = StreamingChatbotImpl{}
)

func getOpenAIClient() *openai.Client {
	once.Do(func() { openAIClient = openai.NewClient(os.Getenv("OPENAI_TOKEN")) })
	return openAIClient
}

func GetChatbot(streaming bool) Chatbot {
	if streaming {
		return &streamingChatbot
	}
	return &oneshotChatbot
}

func populatePromptVariables(prompt string, variables map[string]string) string {
	for key := range variables {
		prompt = strings.ReplaceAll(prompt, "$"+key, variables[key])
	}
	return prompt
}

func (*ChatbotImpl) handleRequest(ctx context.Context, session *model.SessionDTO) (<-chan string, error) {
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

	resCh := make(chan string)
	defer close(resCh)
	for _, toolCall := range resp.Choices[0].Message.ToolCalls {
		var args map[string]string
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		res := toolExecutionMap[toolCall.Function.Name](ctx, session, args)
		resCh <- res
	}

	resCh <- resp.Choices[0].Message.Content
	return resCh, nil
}

func (*StreamingChatbotImpl) handleRequest(ctx context.Context, session *model.SessionDTO) (<-chan string, error) {
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

	resCh := make(chan string)
	go func() {
		defer close(resCh)

		toolCalls := make([]openai.FunctionCall, 0)
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
			resCh <- content

			// Gather delta tool calls
			processStreamToolCalls(&toolCalls, res)
		}

		// If tool call was triggered, send over the response
		var args map[string]string
		for _, myToolCall := range toolCalls {
			json.Unmarshal([]byte(myToolCall.Arguments), &args)
			res := toolExecutionMap[myToolCall.Name](ctx, session, args)
			resCh <- res
		}
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
	// utils.Log(ctx, fmt.Sprintf("ChatCompletion request: %+v\n", createRequest(prompt, session, workflowExecCtx, false)))
	ctx, span := middleware.Tracer.Start(ctx, "gptRequest")
	defer span.End()
	stream, err := client.CreateChatCompletionStream(ctx, createOpenAIRequest(prompt, session, workflowExecCtx, true))
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("ChatCompletion error: %v\n", err))
		return nil, err
	}

	// utils.Log(ctx, fmt.Sprintf("ChatCompletion response: %+v\n", resp))
	return stream, nil
}
