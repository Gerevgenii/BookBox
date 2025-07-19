package repository

import (
	"context"
	"time"

	"github.com/project/library/internal/entity"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=../../../generated/mocks/repository_mock.go -package=mocks
type (
	AuthorRepository interface {
		CreateAuthor(ctx context.Context, author entity.Author) (entity.Author, error)
		UpdateAuthor(ctx context.Context, author entity.Author) error
		GetAuthorBooks(ctx context.Context, authorID string) ([]entity.Book, error)
		GetAuthorInfo(ctx context.Context, authorID string) (entity.Author, error)
	}

	BooksRepository interface {
		CreateBook(ctx context.Context, book entity.Book) (entity.Book, error)
		GetBook(ctx context.Context, bookID string) (entity.Book, error)
		UpdateBook(ctx context.Context, book entity.Book) error
	}

	OutboxRepository interface {
		SendMessage(ctx context.Context, idempotencyKey string, kind OutboxKind, message []byte) error
		GetMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]OutboxData, error)
		MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error
	}

	OutboxData struct {
		IdempotencyKey string
		Kind           OutboxKind
		RawData        []byte
	}
)

type OutboxKind int

const (
	OutboxKindUndefined OutboxKind = iota
	OutboxKindBook
	OutboxKindAuthor
)

func (o OutboxKind) String() string {
	switch o {
	case OutboxKindBook:
		return "book"
	case OutboxKindAuthor:
		return "author"
	default:
		return "undefined"
	}
}
