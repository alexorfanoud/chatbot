package dao

import (
	"chat/internal/data/db"
	store "chat/internal/data/store/distrib"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
)

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
