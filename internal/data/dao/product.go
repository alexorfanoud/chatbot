package dao

import (
	"chat/internal/data/db"
	"chat/internal/model"
	"context"
)

func GetProductByID(ctx context.Context, id int) (*model.Product, error) {

	const stmt = `SELECT id, name, description, price FROM products WHERE id = ?`
	row := db.QueryRowContext(ctx, stmt, id)

	e := &model.Product{}
	err := row.Scan(&e.ID, &e.Name, &e.Description, &e.Price)
	if err != nil {
		return &model.Product{}, err
	}

	return e, nil
}
