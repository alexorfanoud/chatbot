package dao

import (
	"chat/internal/data/db"
	"chat/internal/model"
	"context"
)

func CreateConversation(ctx context.Context, conversation *model.ConversationDTO) error {
	const stmt = `INSERT INTO conversations(workflow, session_id) values(?, ?)`
	res, err := db.ExecContext(ctx, stmt, conversation.WFExecutionContext.Workflow, conversation.SessionID)
	conversation.ID, err = res.LastInsertId()
	return err
}

func CreateActiveConversationForSession(ctx context.Context, session *model.SessionDTO, wfExecCtx *model.WorkflowExecutionContext) (*model.ConversationDTO, error) {
	conv := &model.ConversationDTO{SessionID: session.ID, WFExecutionContext: wfExecCtx, ConversationHistory: make([]*model.Request, 0)}
	err := CreateConversation(ctx, conv)
	if err != nil {
		return nil, err
	}
	// Append it and set it as active
	session.Conversations = append(session.Conversations, conv)
	session.ActiveConversationIdx = len(session.Conversations) - 1

	return conv, nil
}
