package persistence

import (
	"context"
	"database/sql"
	"fmt"
)

// ReadOnlyExecutor runs AI-generated SQL on a mode=ro SQLite connection,
// so even a validation slip cannot mutate data.
type ReadOnlyExecutor struct {
	db *sql.DB
}

func NewReadOnlyExecutor(path string) (*ReadOnlyExecutor, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open read-only sqlite: %w", err)
	}
	return &ReadOnlyExecutor{db: db}, nil
}

func (e *ReadOnlyExecutor) Close() error { return e.db.Close() }

func (e *ReadOnlyExecutor) Query(ctx context.Context, query string) ([]string, [][]string, error) {
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	var out [][]string
	for rows.Next() {
		raw := make([]sql.NullString, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(cols))
		for i, v := range raw {
			if v.Valid {
				row[i] = v.String
			} else {
				row[i] = "NULL"
			}
		}
		out = append(out, row)
	}
	return cols, out, rows.Err()
}
