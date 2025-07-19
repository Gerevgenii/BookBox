package repository

import (
	"context"
	"slices"
	"sync"

	"github.com/project/library/internal/entity"
)

var _ AuthorRepository = (*inMemoryImpl)(nil)
var _ BooksRepository = (*inMemoryImpl)(nil)

type inMemoryImpl struct {
	authorsMx *sync.RWMutex
	authors   map[string]*entity.Author

	booksMx *sync.RWMutex
	books   map[string]*entity.Book
}

func (i *inMemoryImpl) UpdateBook(_ context.Context, book entity.Book) error {
	i.booksMx.Lock()
	defer i.booksMx.Unlock()
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()
	for _, author := range book.AuthorIDs {
		if _, ok := i.authors[author]; !ok {
			return entity.ErrAuthorNotFound
		}
	}

	i.books[book.ID] = &book
	return nil
}

func (i *inMemoryImpl) UpdateAuthor(_ context.Context, author entity.Author) error {
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()
	i.authors[author.ID] = &author
	return nil
}

func (i *inMemoryImpl) GetAuthorBooks(_ context.Context, authorID string) ([]entity.Book, error) {
	i.booksMx.Lock()
	defer i.booksMx.Unlock()
	res := make([]entity.Book, 0)
	for _, book := range i.books {
		if slices.Contains(book.AuthorIDs, authorID) {
			res = append(res, *book)
		}
	}
	return res, nil
}

func (i *inMemoryImpl) GetAuthorInfo(_ context.Context, authorID string) (entity.Author, error) {
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()
	author, ok := i.authors[authorID]
	if !ok {
		return entity.Author{}, entity.ErrAuthorNotFound
	}
	return *author, nil
}

func NewInMemoryRepository() *inMemoryImpl {
	return &inMemoryImpl{
		authorsMx: new(sync.RWMutex),
		authors:   make(map[string]*entity.Author),

		books:   map[string]*entity.Book{},
		booksMx: new(sync.RWMutex),
	}
}

func (i *inMemoryImpl) CreateAuthor(_ context.Context, author entity.Author) (entity.Author, error) {
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()

	if _, ok := i.authors[author.ID]; ok {
		return entity.Author{}, entity.ErrAuthorAlreadyExists
	}

	i.authors[author.ID] = &author
	return author, nil
}

func (i *inMemoryImpl) CreateBook(_ context.Context, book entity.Book) (entity.Book, error) {
	i.booksMx.Lock()
	defer i.booksMx.Unlock()
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()

	if _, ok := i.books[book.ID]; ok {
		return entity.Book{}, entity.ErrBookAlreadyExists
	}

	for _, author := range book.AuthorIDs {
		if _, ok := i.authors[author]; !ok {
			return entity.Book{}, entity.ErrAuthorNotFound
		}
	}
	i.books[book.ID] = &book
	return book, nil
}

func (i *inMemoryImpl) GetBook(_ context.Context, bookID string) (entity.Book, error) {
	i.booksMx.RLock()
	defer i.booksMx.RUnlock()
	v, ok := i.books[bookID]
	if !ok {
		return entity.Book{}, entity.ErrBookNotFound
	}
	return *v, nil
}
