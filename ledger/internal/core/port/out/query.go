package out

import "context"

// QueryExecutor runs raw SELECT queries on a read-only connection.
type QueryExecutor interface {
	Query(ctx context.Context, sql string) (cols []string, rows [][]string, err error)
}
