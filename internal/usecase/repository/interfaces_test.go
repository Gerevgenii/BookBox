package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	t.Parallel()
	const (
		outboxKindBook   = "book"
		outboxKindAuthor = "author"
		undefined        = "undefined"
	)
	require.Equal(t, outboxKindBook, OutboxKindBook.String())
	require.Equal(t, outboxKindAuthor, OutboxKindAuthor.String())
	require.Equal(t, undefined, OutboxKindUndefined.String())
}
