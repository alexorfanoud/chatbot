package conversation

import (
	"chat/internal/data/dao"
	store "chat/internal/data/store/distrib"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"fmt"
)

var chatbot = ChatbotImpl{}

type ConversationManager interface {
	HandleRequestStreaming(context.Context, model.ConversationRequest) (<-chan string, error)
}
type ConversationManagerImpl struct{}

func (cm *ConversationManagerImpl) HandleRequestStreaming(ctx context.Context, cr model.ConversationRequest) (<-chan string, error) {
	// get user session
	session, err := dao.GetOrCreateUserSession(ctx, cr.UserID, cr.ChannelType)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to retrieve session for user id: %d\n, error: %s", cr.UserID, err.Error()))
		return nil, err
	}
	utils.Log(ctx, fmt.Sprintf("Retrieved session: %+v", session))

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
			return nil, err
		}
	}
	utils.Log(ctx, fmt.Sprintf("Retrieved active conversation: %+v", activeConv))

	// create and store the request model
	convRequest := model.Request{ConversationID: activeConv.ID, Question: cr.Request}
	activeConv.AddRequestToHistory(&convRequest)

	// 4. Chatbot request
	chatbotResCh, err := chatbot.handleRequestStreaming(ctx, session)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to get response from chatbot: %s", err.Error()))
	}

	resChan := make(chan string)
	go func() {
		defer close(resChan)
		// store request / session context in cache when processing is done
		defer dao.CreateRequest(ctx, &convRequest)
		defer store.StoreSession(ctx, session)
		for {
			data, ok := <-chatbotResCh
			if !ok {
				break
			}

			if data.IsEOF {
				ans := FulfillWorkflows(ctx, session, data.WorkflowFulfillments)
				// Send the total answer content
				convRequest.Answer = data.Content
				resChan <- ans
				// So that ui can render next chunk as a new msg
				resChan <- "_"
				break
			}

			// Send the delta contents
			resChan <- data.Content
		}
	}()

	return resChan, err
}
