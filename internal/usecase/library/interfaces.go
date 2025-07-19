package library

import (
	"context"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=../../../generated/mocks/usecase_mock.go -package=mocks
type (
	AuthorUseCase interface {
		RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error)
		UpdateAuthor(ctx context.Context, authorID string, authorName string) error
		GetAuthorBooks(ctx context.Context, authorID string) ([]entity.Book, error)
		GetAuthorInfo(ctx context.Context, authorID string) (entity.Author, error)
	}

	BooksUseCase interface {
		RegisterBook(ctx context.Context, name string, authorIDs []string) (entity.Book, error)
		GetBook(ctx context.Context, bookID string) (entity.Book, error)
		UpdateBook(ctx context.Context, bookID string, bookName string, authorIDs []string) error
	}
)

var _ AuthorUseCase = (*libraryImpl)(nil)
var _ BooksUseCase = (*libraryImpl)(nil)

type libraryImpl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	booksRepository  repository.BooksRepository
	outboxRepository repository.OutboxRepository
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	authorRepository repository.AuthorRepository,
	booksRepository repository.BooksRepository,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) *libraryImpl {
	return &libraryImpl{
		logger:           logger,
		authorRepository: authorRepository,
		booksRepository:  booksRepository,
		outboxRepository: outboxRepository,
		transactor:       transactor,
	}
}
