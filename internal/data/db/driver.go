package db

import (
	"chat/internal/api/middleware"
	"context"
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var dbInternal *sql.DB

func InitDB() error {
	db, err := sql.Open("mysql", "root:pass@tcp(localhost:3306)/vendor1")
	dbInternal = db
	if err != nil {
		return err
	}
	return nil
}

func CloseDB() error {
	return dbInternal.Close()
}

func ExecContext(ctx context.Context, stmt string, args ...any) (sql.Result, error) {
	ctx, span := middleware.Tracer.Start(ctx, "dbRequest", trace.WithAttributes(
		attribute.String("type", strings.Split(stmt, " ")[0]),
		attribute.String("stmt", stmt),
	))
	defer span.End()
	return dbInternal.ExecContext(ctx, stmt, args...)
}

func QueryRowContext(ctx context.Context, stmt string, args ...any) *sql.Row {
	ctx, span := middleware.Tracer.Start(ctx, "dbRequest", trace.WithAttributes(
		attribute.String("type", strings.Split(stmt, " ")[0]),
		attribute.String("stmt", stmt),
	))
	defer span.End()
	return dbInternal.QueryRowContext(ctx, stmt, args...)
}

func QueryRowsContext(ctx context.Context, stmt string, args ...any) (*sql.Rows, error) {
	ctx, span := middleware.Tracer.Start(ctx, "dbRequest", trace.WithAttributes(
		attribute.String("type", strings.Split(stmt, " ")[0]),
		attribute.String("stmt", stmt),
	))
	defer span.End()
	return dbInternal.QueryContext(ctx, stmt, args...)
}
