package conversation

import (
	"chat/internal/data/dao"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"fmt"
	"strconv"
	"strings"
)

var (
	workflowFulfillmentMap = map[string]func(context.Context, *model.SessionDTO, map[string]string) string{
		"submit_review": func(ctx context.Context, s *model.SessionDTO, args map[string]string) string {
			return SubmitProductReview(ctx, s, args)
		},
		"complete_return": func(ctx context.Context, s *model.SessionDTO, args map[string]string) string {
			return CompleteProductReturn(ctx, s, args)
		},
	}
)

func FulfillWorkflow(ctx context.Context, session *model.SessionDTO, wfFulfill model.WorkflowFulfillment) string {
	if fulfillment := workflowFulfillmentMap[wfFulfill.Name]; fulfillment != nil {
		return fulfillment(ctx, session, wfFulfill.Arguments)
	}

	return ""
}

func FulfillWorkflows(ctx context.Context, session *model.SessionDTO, workflowFulfillments []model.WorkflowFulfillment) string {
	res := ""
	for _, wfFulfill := range workflowFulfillments {
		if fulfillment := workflowFulfillmentMap[wfFulfill.Name]; fulfillment != nil {
			res += fulfillment(ctx, session, wfFulfill.Arguments)
		}
	}

	return res
}

func CheckWorkflowTriggered(ctx context.Context, session *model.SessionDTO, question string) *model.WorkflowExecutionContext {
	var workflowCtx model.WorkflowExecutionContext
	// Should have some kind of keyword based classification model instead
	if strings.Contains(question, "return") {
		workflowCtx = model.WorkflowExecutionContext{Workflow: model.RETURN,
			ContextWindow:   5,
			PromptVariables: map[string]string{"user_id": strconv.FormatInt(session.UserID, 10)}}
	} else {
		workflowCtx = model.WorkflowExecutionContext{Workflow: model.UNKNOWN}
	}

	return &workflowCtx
}

func CompleteProductReturn(ctx context.Context, s *model.SessionDTO, params map[string]string) string {
	utils.Log(ctx, fmt.Sprintf("Completing return: %+v", params))

	// Conversation is complete
	s.ActiveConversationIdx = -1
	// NOOP

	return "Great! I will get the return request started! What else can i do for you today?"
}

func SubmitProductReview(ctx context.Context, s *model.SessionDTO, params map[string]string) string {
	utils.Log(ctx, fmt.Sprintf("Submitting review: %+v", params))

	pvars := s.GetActiveConversation().WFExecutionContext.PromptVariables
	rstarsi, _ := strconv.Atoi(params["review_stars"])
	uidi, _ := strconv.Atoi(pvars["user_id"])
	pidi, _ := strconv.Atoi(pvars["product_id"])
	// Conversation is complete
	s.ActiveConversationIdx = -1
	err := dao.InsertReview(ctx, &model.Review{Stars: rstarsi, UserID: uidi, ProductId: pidi})
	if err != nil {
		panic(err)
	}

	return "Perfect, Thank you for the review! Let me know if there is anything else I can help with!"
}
