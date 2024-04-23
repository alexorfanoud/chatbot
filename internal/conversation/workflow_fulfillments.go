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

	return "Great! Thank you for the review. Let me know if there is anything else I can help with!"
}
