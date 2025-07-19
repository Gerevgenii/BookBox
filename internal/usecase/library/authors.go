package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/internal/usecase/repository"

	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error) {
	var author entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		author, txErr = l.authorRepository.CreateAuthor(ctx, entity.Author{
			Name: authorName,
		})

		if txErr != nil {
			return txErr
		}

		serialized, txErr := json.Marshal(author)

		if txErr != nil {
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		txErr = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)

		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return entity.Author{}, err
	}

	return author, nil
}

func (l *libraryImpl) UpdateAuthor(ctx context.Context, authorID string, authorName string) error {
	err := l.authorRepository.UpdateAuthor(ctx, entity.Author{
		ID:   authorID,
		Name: authorName,
	})

	if err != nil {
		return err
	}

	return nil
}

func (l *libraryImpl) GetAuthorBooks(ctx context.Context, authorID string) ([]entity.Book, error) {
	author, err := l.authorRepository.GetAuthorBooks(ctx, authorID)

	if err != nil {
		return []entity.Book{}, err
	}

	return author, nil
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, authorID string) (entity.Author, error) {
	author, err := l.authorRepository.GetAuthorInfo(ctx, authorID)

	if err != nil {
		return entity.Author{}, err
	}

	return author, nil
}
