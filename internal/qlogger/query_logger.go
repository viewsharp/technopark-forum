package qlogger

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

var logger = zap.Must(zap.NewProduction())

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type QueryLogger struct {
	db DB
}

func NewQueryLogger(db DB) *QueryLogger {
	return &QueryLogger{db: db}
}

func (q *QueryLogger) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	startedAt := time.Now()

	row := q.db.QueryRow(ctx, sql, args...)

	logger.Info("query row",
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("query", sql),
		//zap.Any("args", args),
	)

	return row
}

func (q *QueryLogger) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	startedAt := time.Now()

	rows, err := q.db.Query(ctx, sql, args...)

	logger.Info("query",
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("query", sql),
		//zap.Any("args", args),
	)

	return rows, err
}

func (q *QueryLogger) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	startedAt := time.Now()

	result, err := q.db.Exec(ctx, sql, arguments...)

	logger.Info("exec",
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("query", sql),
		//zap.Any("args", arguments),
	)

	return result, err
}
