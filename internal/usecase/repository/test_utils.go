package repository

import (
	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
)

func CreateBook(name string, authorsID ...string) entity.Book {
	return entity.Book{
		ID:        uuid.New().String(),
		Name:      name,
		AuthorIDs: authorsID,
	}
}

func CreateAuthor(name string) entity.Author {
	return entity.Author{
		ID:   uuid.New().String(),
		Name: name,
	}
}
