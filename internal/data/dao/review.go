package dao

import (
	"chat/internal/data/db"
	"chat/internal/model"
	"context"
)

func InsertReview(ctx context.Context, review *model.Review) error {

	const stmt = `INSERT INTO reviews(user_id, product_id, score) values(?, ?, ?)`
	_, err := db.ExecContext(ctx, stmt, review.UserID, review.ProductId, review.Stars)

	if err != nil {
		return err
	}

	return nil
}
