package db

import (
	"context"
)

func (q *Queries) EstimateRowCount(ctx context.Context, tableName string) (int64, error) {
	const query = `
		SELECT COALESCE(reltuples::bigint, 0)
		FROM pg_class
		WHERE relname = $1
	`
	var count int64
	err := q.db.QueryRow(ctx, query, tableName).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}