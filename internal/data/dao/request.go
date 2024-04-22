package dao

import (
	"chat/internal/data/db"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
)

func CreateRequest(ctx context.Context, request *model.Request) error {
	const stmt = `INSERT INTO requests(conversation_id, question, answer) values(?, ?, ?)`
	res, err := db.ExecContext(ctx, stmt, request.ConversationID, request.Question, request.Answer)
	if err != nil {
		utils.Log(ctx, "Failed to insert request to db")
		return err
	}
	request.ID, err = res.LastInsertId()
	return err
}

func UpdateRequest(ctx context.Context, request *model.Request) error {
	const stmt = `UPDATE requests set conversation_id=?, question=?, answer=? WHERE id =?`
	_, err := db.ExecContext(ctx, stmt, request.ConversationID, request.Question, request.Answer, request.ID)
	if err != nil {
		utils.Log(ctx, "Failed to update request in db")
		return err
	}
	return err
}
