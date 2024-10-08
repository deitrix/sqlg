package sqlg

import (
	"context"
	"database/sql"
	"log/slog"
)

var Debug bool

type Queryable interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func SelectAll[T any, Dataset dataset[Dataset]](ctx context.Context, q Queryable, scan ScanFunc[T], ds Dataset) ([]T, error) {
	sql, args, err := ds.Prepared(!Debug).ToSQL()
	if err != nil {
		return nil, err
	}
	if Debug {
		slog.Debug(sql)
	}
	rows, err := q.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows, scan)
}

func Select[T any, Dataset dataset[Dataset]](ctx context.Context, q Queryable, scan ScanFunc[T], ds Dataset) (T, error) {
	sql, args, err := ds.Prepared(!Debug).ToSQL()
	if err != nil {
		var zero T
		return zero, err
	}
	if Debug {
		slog.Debug(sql)
	}
	row := q.QueryRowContext(ctx, sql, args...)
	return scan(row)
}

type dataset[T any] interface {
	ToSQL() (string, []interface{}, error)
	Prepared(bool) T
}

func Exec[T dataset[T]](ctx context.Context, q Queryable, ds T) error {
	_, err := ExecResult(ctx, q, ds)
	return err
}

func ExecResult[Dataset dataset[Dataset]](ctx context.Context, q Queryable, ds Dataset) (sql.Result, error) {
	sql, args, err := ds.Prepared(!Debug).ToSQL()
	if err != nil {
		return nil, err
	}
	if Debug {
		slog.Debug(sql)
	}
	return q.ExecContext(ctx, sql, args...)
}

func ExecID[Dataset dataset[Dataset]](ctx context.Context, q Queryable, ds Dataset) (int, error) {
	result, err := ExecResult(ctx, q, ds)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func ExecRowsAffected[Dataset dataset[Dataset]](ctx context.Context, q Queryable, ds Dataset) (int, error) {
	result, err := ExecResult(ctx, q, ds)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

type Row interface {
	Scan(dest ...any) error
}

type ScanFunc[T any] func(Row) (T, error)

func scanRows[T any](rows *sql.Rows, f ScanFunc[T]) ([]T, error) {
	var results []T
	for rows.Next() {
		result, err := f(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
