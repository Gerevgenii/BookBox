package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

//go:generate ../../../bin/mockgen -source=transactor.go -destination=../../../generated/mocks/transactor_mock.go -package=mocks
type Transactor interface {
	WithTx(ctx context.Context, function func(ctx context.Context) error) error
}

var _ Transactor = (*transactorImpl)(nil)

type transactorImpl struct {
	db MyPgxPool
}

func NewTransactor(db MyPgxPool) *transactorImpl {
	return &transactorImpl{
		db: db,
	}
}

func (t *transactorImpl) WithTx(ctx context.Context, function func(ctx context.Context) error) (txErr error) {
	ctxWithTx, tx, err := injectTx(ctx, t.db)

	if err != nil {
		return err
	}

	defer func() {
		if txErr != nil {
			newErr := tx.Rollback(ctxWithTx)
			if newErr != nil {
				err = newErr
			}
			return
		}

		newErr := tx.Commit(ctxWithTx)
		if newErr != nil {
			err = newErr
		}
	}()

	if err != nil {
		return err
	}

	err = function(ctxWithTx)

	if err != nil {
		return err
	}

	return nil
}

type txInjector struct{}

var ErrTxNotFound = errors.New("tx not found in context")

func injectTx(ctx context.Context, pool MyPgxPool) (context.Context, pgx.Tx, error) {
	if tx, err := extractTx(ctx); err == nil {
		return ctx, tx, nil
	}

	tx, err := pool.Begin(ctx)

	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, txInjector{}, tx), tx, nil
}

func extractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txInjector{}).(pgx.Tx)

	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}
