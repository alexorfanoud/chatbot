package dao

import (
	"chat/internal/data/db"
	"chat/internal/model"
	"context"
)

func GetUserById(ctx context.Context, userId int) (*model.User, error) {

	const stmt = `SELECT id, name, email, phone FROM users WHERE id = ?`
	row := db.QueryRowContext(ctx, stmt, userId)

	user := &model.User{}
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Phone)
	if err != nil {
		return &model.User{}, err
	}

	return user, nil
}
