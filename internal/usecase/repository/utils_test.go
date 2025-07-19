package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func getPgxMockPool(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	newPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() {
		newPool.Close()
	})
	return newPool
}

func TestMyExtractCtx(t *testing.T) {
	t.Parallel()
	successString := "Hello World"
	tests := []struct {
		name        string
		pool        pgxmock.PgxPoolIface
		function    func(tx pgx.Tx) (string, error)
		ctx         context.Context
		expectedStr string
	}{
		{
			name: "success test",
			pool: func() pgxmock.PgxPoolIface {
				newPool := getPgxMockPool(t)
				newPool.ExpectBegin()
				return newPool
			}(),
			function: func(tx pgx.Tx) (string, error) {
				return successString, nil
			},
			expectedStr: successString,
			ctx:         t.Context(),
		},
		{
			name: "failure pool begin",
			pool: func() pgxmock.PgxPoolIface {
				newPool := getPgxMockPool(t)
				newPool.ExpectBegin().WillReturnError(errors.New("error"))
				return newPool
			}(),
			function: func(tx pgx.Tx) (string, error) {
				return successString, nil
			},
			expectedStr: "",
			ctx:         t.Context(),
		},
		{
			name: "failure function",
			pool: func() pgxmock.PgxPoolIface {
				newPool := getPgxMockPool(t)
				newPool.ExpectBegin()
				return newPool
			}(),
			function: func(tx pgx.Tx) (string, error) {
				return "", errors.New("error")
			},
			expectedStr: "",
			ctx:         t.Context(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			str, err := myExtractCtx(test.ctx, test.pool, test.function)
			if str != "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, test.expectedStr, str)
		})
	}
}

func TestMyExtractCtxNoT(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		pool        pgxmock.PgxPoolIface
		function    func(tx pgx.Tx) error
		ctx         context.Context
		expectedStr string
	}{
		{
			name: "success test",
			pool: func() pgxmock.PgxPoolIface {
				newPool := getPgxMockPool(t)
				newPool.ExpectBegin()
				return newPool
			}(),
			function: func(tx pgx.Tx) error {
				return nil
			},
			expectedStr: "Hello World",
			ctx:         t.Context(),
		},
		{
			name: "failure pool begin",
			pool: func() pgxmock.PgxPoolIface {
				newPool := getPgxMockPool(t)
				newPool.ExpectBegin().WillReturnError(errors.New("error"))
				return newPool
			}(),
			function: func(tx pgx.Tx) error {
				return nil
			},
			expectedStr: "",
			ctx:         t.Context(),
		},
		{
			name: "failure function",
			pool: func() pgxmock.PgxPoolIface {
				newPool := getPgxMockPool(t)
				newPool.ExpectBegin()
				return newPool
			}(),
			function: func(tx pgx.Tx) error {
				return errors.New("error")
			},
			expectedStr: "",
			ctx:         t.Context(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := myExtractCtxNoT(test.ctx, test.pool, test.function)
			if test.expectedStr != "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
