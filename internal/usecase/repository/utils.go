package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func myExtractCtx[T any](ctx context.Context, pgxPool MyPgxPool, function func(tx pgx.Tx) (T, error)) (res T, txErr error) {
	var (
		tx  pgx.Tx
		err error
	)

	if tx, err = extractTx(ctx); err != nil {
		tx, err = pgxPool.Begin(ctx)

		if err != nil {
			var ans T
			return ans, err
		}

		defer func(tx pgx.Tx, ctx context.Context) {
			if txErr != nil {
				newErr := tx.Rollback(ctx)
				if newErr != nil {
					err = newErr
				}
				return
			}
			newErr := tx.Commit(ctx)
			if newErr != nil {
				err = newErr
			}
		}(tx, ctx)
	}

	return function(tx)
}

func myExtractCtxNoT(ctx context.Context, pgxPool MyPgxPool, function func(tx pgx.Tx) error) (txErr error) {
	var (
		tx  pgx.Tx
		err error
	)

	if tx, err = extractTx(ctx); err != nil {
		tx, err = pgxPool.Begin(ctx)

		if err != nil {
			return err
		}

		defer func(tx pgx.Tx, ctx context.Context) {
			if txErr != nil {
				newErr := tx.Rollback(ctx)
				if newErr != nil {
					err = newErr
				}
				return
			}
			newErr := tx.Commit(ctx)
			if newErr != nil {
				err = newErr
			}
		}(tx, ctx)
	}

	return function(tx)
}
