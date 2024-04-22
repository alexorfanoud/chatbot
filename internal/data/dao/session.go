package dao

import (
	"chat/internal/data/db"
	store "chat/internal/data/store/distrib"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
)

func CreateSession(ctx context.Context, session *model.SessionDTO) error {
	const stmt = `INSERT INTO sessions(user_id) values(?)`
	res, err := db.ExecContext(ctx, stmt, session.UserID)
	if err != nil {
		return err
	}
	session.ID, err = res.LastInsertId()

	return err
}

func GetOrCreateUserSession(ctx context.Context, uid int64) (*model.SessionDTO, error) {
	// session context is stored in redis
	session, err := store.GetSession(ctx, uid)
	if err != nil {
		utils.Log(ctx, "Unable to retrieve store from store")
		return nil, err
	}

	if session == nil {
		session = &model.SessionDTO{UserID: uid, Conversations: []*model.ConversationDTO{}, ActiveConversationIdx: -1}
		err = CreateSession(ctx, session)
		if err != nil {
			utils.Log(ctx, "Unable to retrieve store from store")
			return nil, err
		}
	}

	return session, nil
}
