package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/internal/usecase/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterBook(ctx context.Context, name string, authorIDs []string) (entity.Book, error) {
	span := trace.SpanFromContext(ctx)
	l.logger.Info("start to register book",
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	var book entity.Book

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		book, txErr = l.booksRepository.CreateBook(ctx, entity.Book{
			Name:      name,
			AuthorIDs: authorIDs,
		})

		if txErr != nil {
			return txErr
		}

		serialized, txErr := json.Marshal(book)

		if txErr != nil {
			return txErr
		}

		idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID
		txErr = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindBook, serialized)

		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		return entity.Book{}, err
	}

	span.SetAttributes(attribute.String("book.id", book.ID))
	l.logger.Info("book registered",
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("book_id", book.ID),
	)

	return book, nil
}

func (l *libraryImpl) GetBook(ctx context.Context, bookID string) (entity.Book, error) {
	return l.booksRepository.GetBook(ctx, bookID)
}

func (l *libraryImpl) UpdateBook(ctx context.Context, bookID string, bookName string, authorIDs []string) error {
	return l.booksRepository.UpdateBook(ctx, entity.Book{
		ID:        bookID,
		Name:      bookName,
		AuthorIDs: authorIDs,
	})
}
