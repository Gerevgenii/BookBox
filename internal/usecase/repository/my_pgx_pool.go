package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

type MyPgxPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type MyPgxPoolSmart struct {
	pool pgxmock.PgxPoolIface
}

func (p *MyPgxPoolSmart) Begin(_ context.Context) (pgx.Tx, error) {
	return p.pool, nil
}

type MyPgxPoolDump struct {
	err error
}

func (p *MyPgxPoolDump) Begin(_ context.Context) (pgx.Tx, error) {
	return nil, p.err
}

type MyPgxOutboxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}
