package repository

import (
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestNewOutbox(t *testing.T) {
	t.Parallel()
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	outbox := NewOutbox(pool)
	require.NotNil(t, outbox)
}
