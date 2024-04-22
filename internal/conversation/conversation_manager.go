package conversation

import (
	"chat/internal/data/dao"
	store "chat/internal/data/store/distrib"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"fmt"
)

type ConversationManager interface {
	HandleRequest()
}

type ConversationManagerImpl struct{}

var chatbot = ChatbotImpl{}

func (*ConversationManagerImpl) HandleRequest(ctx context.Context, question string, userId int64, startNewConv bool, workflowExecCtx *model.WorkflowExecutionContext) (string, error) {
	// get user session
	session, err := dao.GetOrCreateUserSession(ctx, userId)
	utils.Log(ctx, fmt.Sprintf("Retrieved session: %+v", session))
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to retrieve session for user id: %d\n, error: %s", userId, err.Error()))
		return "", err
	}
	// store session / conversation context in cache when processing is done
	defer func() {
		err := store.StoreSession(ctx, int(userId), session)
		if err != nil {
			utils.Log(ctx, fmt.Sprintf("Unable to store session for user: %d - %s", userId, err.Error()))
		}
	}()

	// check if new workflow should be triggered based on the request
	discoveredWorkflowCtx := CheckWorkflowTriggered(ctx, session, question)
	if discoveredWorkflowCtx.Workflow != model.UNKNOWN && discoveredWorkflowCtx.Workflow != workflowExecCtx.Workflow {
		workflowExecCtx = discoveredWorkflowCtx
		startNewConv = true
	}

	// get active activeConv (=workflow execution) within the session
	var activeConv *model.ConversationDTO = session.GetActiveConversation()
	if startNewConv || activeConv == nil {
		activeConv, err = dao.CreateActiveConversationForSession(ctx, session, workflowExecCtx)
		if err != nil {
			utils.Log(ctx, fmt.Sprintf("Unable to create new active conversation: %s", err.Error()))
			return "", err
		}
	}
	utils.Log(ctx, fmt.Sprintf("Retrieved active conversation: %+v", activeConv))

	// create and store the request model
	convRequest := model.Request{ConversationID: activeConv.ID, Question: question}
	err = dao.CreateRequest(ctx, &convRequest)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to insert request: %s, error: %s", question, err.Error()))
		return "", err
	}
	activeConv.AddRequestToHistory(&convRequest)

	// 4. Chatbot request
	chatbotResp, err := chatbot.handleRequest(ctx, session)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to get chatbot response, error: %s", err.Error()))
		return "", err
	}

	// 5. Update request answer
	convRequest.Answer = chatbotResp
	go dao.UpdateRequest(ctx, &convRequest)
	return chatbotResp, err
}
