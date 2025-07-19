package repository

import (
	"github.com/project/library/internal/entity"

	"testing"

	"github.com/stretchr/testify/require"
)

func booksRepository(t *testing.T, target *inMemoryImpl, books ...entity.Book) *inMemoryImpl {
	t.Helper()
	for _, book := range books {
		_, err := target.CreateBook(t.Context(), book)
		require.NoError(t, err)
	}
	return target
}

func authorRepository(t *testing.T, target *inMemoryImpl, authors ...entity.Author) *inMemoryImpl {
	t.Helper()
	for _, author := range authors {
		_, err := target.CreateAuthor(t.Context(), author)
		require.NoError(t, err)
	}
	return target
}

func createInMemoryRepository(t *testing.T, books []entity.Book, authors []entity.Author) *inMemoryImpl {
	t.Helper()
	target := NewInMemoryRepository()

	authorRepository(t, target, authors...)
	booksRepository(t, target, books...)
	return target
}

func TestCreateAuthor(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
	}
	author := CreateAuthor("Dave")
	target := authorRepository(t, NewInMemoryRepository(), authors...)
	tests := []struct {
		name        string
		target      AuthorRepository
		author      entity.Author
		expected    entity.Author
		expectedErr error
	}{
		{
			name:        "Success case",
			target:      target,
			author:      author,
			expected:    author,
			expectedErr: nil,
		},
		{
			name:        "Author already exists",
			target:      target,
			author:      authors[0],
			expected:    entity.Author{},
			expectedErr: entity.ErrAuthorAlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, actualErr := test.target.CreateAuthor(t.Context(), test.author)
			require.ErrorIs(t, test.expectedErr, actualErr)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestUpdateAuthor(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
	}
	authors[0].Name = "Dave"
	target := authorRepository(t, NewInMemoryRepository(), authors...)
	tests := []struct {
		name        string
		target      AuthorRepository
		author      entity.Author
		expectedErr error
	}{
		{
			name:        "Success case",
			target:      target,
			author:      authors[0],
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualErr := test.target.UpdateAuthor(t.Context(), test.author)
			require.ErrorIs(t, test.expectedErr, actualErr)
		})
	}
}

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
		CreateAuthor("Dave"),
	}
	books := []entity.Book{
		CreateBook("How to live in the beauty trash", authors[0].ID),
		CreateBook("How to not live in the beauty trash", authors[1].ID),
		CreateBook("Live in the beauty", authors[1].ID, authors[2].ID),
	}
	tests := []struct {
		name        string
		target      AuthorRepository
		authorID    string
		expected    []entity.Book
		expectedErr error
	}{
		{
			name:        "Alice case",
			target:      createInMemoryRepository(t, books, authors),
			authorID:    authors[0].ID,
			expected:    []entity.Book{books[0]},
			expectedErr: nil,
		},
		{
			name:        "Bob case",
			target:      createInMemoryRepository(t, books, authors),
			authorID:    authors[1].ID,
			expected:    []entity.Book{books[1], books[2]},
			expectedErr: nil,
		},
		{
			name:        "Charlie case",
			target:      createInMemoryRepository(t, books, authors),
			authorID:    authors[2].ID,
			expected:    []entity.Book{books[2]},
			expectedErr: nil,
		},
		{
			name:        "Dave case",
			target:      createInMemoryRepository(t, books, authors),
			authorID:    authors[3].ID,
			expected:    []entity.Book{},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, actualErr := test.target.GetAuthorBooks(t.Context(), test.authorID)
			require.ErrorIs(t, test.expectedErr, actualErr)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
		CreateAuthor("Dave"),
	}
	target := authorRepository(t, NewInMemoryRepository(), authors...)
	tests := []struct {
		name        string
		target      AuthorRepository
		authorID    string
		expected    entity.Author
		expectedErr error
	}{
		{
			name:        "Success case",
			target:      target,
			authorID:    authors[0].ID,
			expected:    authors[0],
			expectedErr: nil,
		},
		{
			name:        "Failure case",
			target:      target,
			authorID:    "failure",
			expected:    entity.Author{},
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, actualErr := test.target.GetAuthorInfo(t.Context(), test.authorID)
			require.ErrorIs(t, test.expectedErr, actualErr)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestCreateBook(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
		CreateAuthor("Dave"),
	}
	books := []entity.Book{
		CreateBook("How to live in the beauty trash", authors[0].ID),
		CreateBook("How to not live in the beauty trash", authors[1].ID),
		CreateBook("Live in the beauty", authors[1].ID, authors[2].ID),
	}

	book := CreateBook("How to live", authors[0].ID)
	failureBook := CreateBook("Live in the beauty", "failureID")

	tests := []struct {
		name        string
		target      BooksRepository
		book        entity.Book
		expected    entity.Book
		expectedErr error
	}{
		{
			name:        "Success case",
			target:      createInMemoryRepository(t, books, authors),
			book:        book,
			expected:    book,
			expectedErr: nil,
		},
		{
			name:        "Book already exists",
			target:      createInMemoryRepository(t, books, authors),
			book:        books[0],
			expected:    entity.Book{},
			expectedErr: entity.ErrBookAlreadyExists,
		},
		{
			name:        "Author not found",
			target:      createInMemoryRepository(t, books, authors),
			book:        failureBook,
			expected:    entity.Book{},
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, actualErr := test.target.CreateBook(t.Context(), test.book)
			require.ErrorIs(t, test.expectedErr, actualErr)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
		CreateAuthor("Dave"),
	}
	books := []entity.Book{
		CreateBook("How to live in the beauty trash", authors[0].ID),
		CreateBook("How to not live in the beauty trash", authors[1].ID),
		CreateBook("Live in the beauty", authors[1].ID, authors[2].ID),
	}

	books[0].Name = "Something wrong in our life"
	failureBook := CreateBook("Live in the beauty", "failureID")

	tests := []struct {
		name        string
		target      BooksRepository
		book        entity.Book
		expectedErr error
	}{
		{
			name:        "Success case",
			target:      createInMemoryRepository(t, books, authors),
			book:        books[0],
			expectedErr: nil,
		},
		{
			name:        "Author not found",
			target:      createInMemoryRepository(t, books, authors),
			book:        failureBook,
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualErr := test.target.UpdateBook(t.Context(), test.book)
			require.ErrorIs(t, test.expectedErr, actualErr)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()
	authors := []entity.Author{
		CreateAuthor("Alice"),
		CreateAuthor("Bob"),
		CreateAuthor("Charlie"),
		CreateAuthor("Dave"),
	}
	books := []entity.Book{
		CreateBook("How to live in the beauty trash", authors[0].ID),
		CreateBook("How to not live in the beauty trash", authors[1].ID),
		CreateBook("Live in the beauty", authors[1].ID, authors[2].ID),
	}

	failureBook := CreateBook("Live in the beauty", "failureID")

	tests := []struct {
		name        string
		target      BooksRepository
		bookID      string
		expected    entity.Book
		expectedErr error
	}{
		{
			name:        "Success case",
			target:      createInMemoryRepository(t, books, authors),
			bookID:      books[0].ID,
			expected:    books[0],
			expectedErr: nil,
		},
		{
			name:        "Failure case",
			target:      createInMemoryRepository(t, books, authors),
			bookID:      failureBook.ID,
			expected:    entity.Book{},
			expectedErr: entity.ErrBookNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, actualErr := test.target.GetBook(t.Context(), test.bookID)
			require.ErrorIs(t, test.expectedErr, actualErr)
			require.Equal(t, test.expected, actual)
		})
	}
}
