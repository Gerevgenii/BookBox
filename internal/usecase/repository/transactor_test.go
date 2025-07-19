package repository

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestNewTransactor(t *testing.T) {
	t.Parallel()

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)

	transactor := NewTransactor(&MyPgxPoolSmart{
		pool: poolMock,
	})
	require.NotNil(t, transactor)
}

func TestWithTx(t *testing.T) {
	t.Parallel()

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)

	invalidPgxPool := errors.New("invalid pgxpool")
	functionError := errors.New("function error")

	successTarget := NewTransactor(&MyPgxPoolSmart{
		pool: poolMock,
	})
	failureTarget := NewTransactor(&MyPgxPoolDump{
		err: invalidPgxPool,
	})

	tests := []struct {
		name        string
		target      *transactorImpl
		function    func(ctx context.Context) error
		expectedErr error
	}{
		{
			name:   "success test",
			target: successTarget,
			function: func(ctx context.Context) error {
				return nil
			},
			expectedErr: nil,
		},
		{
			name:   "failure test",
			target: failureTarget,
			function: func(ctx context.Context) error {
				return nil
			},
			expectedErr: invalidPgxPool,
		},
		{
			name:   "failure function",
			target: successTarget,
			function: func(ctx context.Context) error {
				return functionError
			},
			expectedErr: functionError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.target.WithTx(t.Context(), test.function)
			if err != nil {
				require.ErrorIs(t, err, test.expectedErr)
			}
		})
	}
}

func TestInjectTx(t *testing.T) {
	t.Parallel()

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)

	successTarget := &MyPgxPoolSmart{
		pool: poolMock,
	}

	ctx := t.Context()

	newCtx, _, err := injectTx(ctx, successTarget)
	require.NoError(t, err)
	newCtx1, _, err1 := injectTx(newCtx, successTarget)
	require.NoError(t, err1)
	require.Equal(t, newCtx, newCtx1)
	require.NoError(t, err)
}
