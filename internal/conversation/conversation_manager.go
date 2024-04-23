package conversation

import (
	"chat/internal/conversation/notification"
	"chat/internal/data/dao"
	store "chat/internal/data/store/distrib"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"fmt"
)

var chatbot = ChatbotImpl{}

type ConversationManager interface {
	HandleRequest(context.Context, model.ConversationRequest) (string, error)
	HandleRequestStreaming(context.Context, model.ConversationRequest) (<-chan string, error)
}
type ConversationManagerImpl struct{}

func (cm *ConversationManagerImpl) HandleRequest(ctx context.Context, cr model.ConversationRequest) (string, error) {
	// get user session
	session, err := dao.GetOrCreateUserSession(ctx, cr.UserID, cr.ChannelType)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to retrieve session for user id: %d\n, error: %s", cr.UserID, err.Error()))
		return "", err
	}
	// store session / conversation context in cache when processing is done
	defer func() {
		err := store.StoreSession(ctx, session)
		if err != nil {
			utils.Log(ctx, fmt.Sprintf("Unable to store session for user: %d - %s", cr.UserID, err.Error()))
		}
	}()

	// check if new workflow should be triggered based on the request
	discoveredWorkflowCtx := CheckWorkflowTriggered(ctx, session, cr.Request)
	if discoveredWorkflowCtx.Workflow != model.UNKNOWN && discoveredWorkflowCtx.Workflow != cr.WorkflowExecutionContext.Workflow {
		cr.WorkflowExecutionContext = discoveredWorkflowCtx
		cr.NewConversation = true
	}

	// get active activeConv (=workflow execution) within the session
	var activeConv *model.ConversationDTO = session.GetActiveConversation()
	if cr.NewConversation || activeConv == nil {
		activeConv, err = dao.CreateActiveConversationForSession(ctx, session, cr.WorkflowExecutionContext)
		if err != nil {
			utils.Log(ctx, fmt.Sprintf("Unable to create new active conversation: %s", err.Error()))
			return "", err
		}
	}
	utils.Log(ctx, fmt.Sprintf("Retrieved active conversation: %+v", activeConv))

	// create and store the request model
	convRequest := model.Request{ConversationID: activeConv.ID, Question: cr.Request}
	activeConv.AddRequestToHistory(&convRequest)
	defer dao.CreateRequest(ctx, &convRequest)

	// 4. Chatbot request
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to get chatbot response, error: %s", err.Error()))
		return "", err
	}

	chatbotResCh, err := GetChatbot(session.ChannelType.SupportsStreaming()).handleRequest(ctx, session)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to get response from chatbot: %s", err.Error()))
	}

	totalResp := ""
	for {
		data, ok := <-chatbotResCh
		if !ok {
			break
		}

		totalResp += data
		notification.GetNotifier(session.ChannelType).Notify(ctx, session.UserID, data)
	}

	// 5. Update request answer
	convRequest.Answer = totalResp
	return "_", err
}
