package qlogger

import (
	"database/sql"
	"time"

	"go.uber.org/zap"
)

var logger = zap.Must(zap.NewProduction())

type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

type QueryLogger struct {
	db DB
}

func NewQueryLogger(db DB) *QueryLogger {
	return &QueryLogger{db: db}
}

func (q *QueryLogger) QueryRow(query string, args ...any) *sql.Row {
	startedAt := time.Now()

	row := q.db.QueryRow(query, args...)

	logger.Info("query row",
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("query", query),
	)

	return row
}

func (q *QueryLogger) Query(query string, args ...any) (*sql.Rows, error) {
	startedAt := time.Now()

	rows, err := q.db.Query(query, args...)

	logger.Info("query",
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("query", query),
	)

	return rows, err
}

func (q *QueryLogger) Exec(query string, args ...any) (sql.Result, error) {
	startedAt := time.Now()

	result, err := q.db.Exec(query, args...)

	logger.Info("exec",
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("query", query),
	)

	return result, err
}
