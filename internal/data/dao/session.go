package dao

import (
	"chat/internal/data/db"
	store "chat/internal/data/store/distrib"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"fmt"
)

func GetRequestsForSession(ctx context.Context, uid int64) (map[int64][]model.Request, error) {
	const stmt = `SELECT r.question, r.answer, r.created_at, c.id, c.workflow from sessions s, conversations c, requests r WHERE s.id = c.session_id and c.id = r.conversation_id and s.user_id = ? ORDER BY r.created_at ASC`
	rows, err := db.QueryRowsContext(ctx, stmt, uid)
	conversations := make(map[int64][]model.Request, 0)

	// Iterate over the rows
	for rows.Next() {
		var request model.Request
		var cid int64
		var workflow int
		if err := rows.Scan(&request.Question, &request.Answer, &request.CreatedAt, &cid, &workflow); err != nil {
			utils.Log(ctx, fmt.Sprintf("Unable to scan row: %s", err.Error()))
		}
		// Append the parsed struct to the slice
		conversations[cid] = append(conversations[cid], request)
	}

	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Unable to get session requests: %s", err.Error()))
		return nil, err
	}

	return conversations, err
}

func CreateSession(ctx context.Context, session *model.SessionDTO) error {
	const stmt = `INSERT INTO sessions(user_id, channel_type) values(?, ?)`
	res, err := db.ExecContext(ctx, stmt, session.UserID, session.ChannelType)
	if err != nil {
		return err
	}
	session.ID, err = res.LastInsertId()

	return err
}

func GetOrCreateUserSession(ctx context.Context, uid int64, ct model.ChannelType) (*model.SessionDTO, error) {
	// session context is stored in redis
	session, err := store.GetSession(ctx, uid, ct)
	if err != nil {
		utils.Log(ctx, "Unable to retrieve session from store")
		return nil, err
	}

	if session == nil {
		session = &model.SessionDTO{UserID: uid, Conversations: []*model.ConversationDTO{}, ActiveConversationIdx: -1, ChannelType: ct}
		err = CreateSession(ctx, session)
		if err != nil {
			utils.Log(ctx, "Unable to retrieve session from store")
			return nil, err
		}
	}

	return session, nil
}
