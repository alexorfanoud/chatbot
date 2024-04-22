package dao

import (
	"chat/internal/data/db"
	"chat/internal/model"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func UpdatePrompt(ctx context.Context, prompt *model.PromptDTO) error {
	if prompt == nil {
		return errors.New("Prompt can not be nil")
	}

	const stmt = `UPDATE prompts set prompt_text = ? WHERE workflow = ? `
	if _, err := db.ExecContext(ctx, stmt, prompt.Text, prompt.Workflow); err != nil {
		return err
	}
	return nil
}

func GetPromptByWorkflow(ctx context.Context, workflow int) (*model.PromptDTO, error) {

	const stmt = `SELECT prompt_text, tools FROM prompts WHERE workflow = ?`
	row := db.QueryRowContext(ctx, stmt, workflow)

	prompt := &model.PromptDTO{}
	var toolsStr sql.NullString
	err := row.Scan(&prompt.Text, &toolsStr)
	if err != nil {
		return &model.PromptDTO{}, err
	}

	err = json.Unmarshal([]byte(toolsStr.String), &prompt.Tools)
	if err != nil {
		return &model.PromptDTO{}, err
	}

	return prompt, nil
}
