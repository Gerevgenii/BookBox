package library

import (
	"context"
	"testing"

	"github.com/project/library/generated/mocks"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

const SUCCESS = "success"
const FAILURE = "failure"

type DumbTransactorImpl struct{}

func (d *DumbTransactorImpl) WithTx(ctx context.Context, function func(ctx context.Context) error) error {
	return function(ctx)
}

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})

	successAuthor := repository.CreateAuthor(SUCCESS)
	failureAuthor := repository.CreateAuthor(FAILURE + "_1")

	authorMock.EXPECT().CreateAuthor(gomock.Any(), gomock.Cond(func(x entity.Author) bool {
		return x.Name == SUCCESS
	})).Return(successAuthor, nil).AnyTimes()
	authorMock.EXPECT().CreateAuthor(gomock.Any(), gomock.Cond(func(x entity.Author) bool {
		return x.Name == FAILURE
	})).Return(entity.Author{}, entity.ErrAuthorAlreadyExists).AnyTimes()
	authorMock.EXPECT().CreateAuthor(gomock.Any(), gomock.Cond(func(x entity.Author) bool {
		return x.Name == FAILURE+"_1"
	})).Return(failureAuthor, nil).AnyTimes()
	outboxMock.EXPECT().SendMessage(gomock.Any(), repository.OutboxKindAuthor.String()+"_"+failureAuthor.ID, repository.OutboxKindAuthor, gomock.Any()).Return(entity.ErrAuthorNotFound).AnyTimes()
	outboxMock.EXPECT().SendMessage(gomock.Any(), gomock.Any(), repository.OutboxKindAuthor, gomock.Any()).Return(nil).AnyTimes()

	tests := []struct {
		name        string
		target      AuthorUseCase
		authorID    string
		expected    entity.Author
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			authorID:    SUCCESS,
			expected:    successAuthor,
			expectedErr: nil,
		},
		{
			name:        "failure case",
			target:      target,
			authorID:    FAILURE,
			expected:    entity.Author{},
			expectedErr: entity.ErrAuthorAlreadyExists,
		},
		{
			name:        "failure send message",
			target:      target,
			authorID:    FAILURE + "_1",
			expected:    entity.Author{},
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := test.target.RegisterAuthor(t.Context(), test.authorID)
			require.ErrorIs(t, test.expectedErr, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestUpdateAuthor(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})

	authorMock.EXPECT().UpdateAuthor(gomock.Any(), gomock.Cond(func(x entity.Author) bool {
		return x.Name == SUCCESS
	})).Return(nil)
	authorMock.EXPECT().UpdateAuthor(gomock.Any(), gomock.Cond(func(x entity.Author) bool {
		return x.Name == FAILURE
	})).Return(entity.ErrAuthorAlreadyExists)

	tests := []struct {
		name        string
		target      AuthorUseCase
		authorID    string
		authorName  string
		expected    entity.Author
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			authorID:    SUCCESS,
			authorName:  SUCCESS,
			expectedErr: nil,
		},
		{
			name:        "failure case",
			target:      target,
			authorID:    FAILURE,
			authorName:  FAILURE,
			expectedErr: entity.ErrAuthorAlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.target.UpdateAuthor(t.Context(), test.authorID, test.authorName)
			require.ErrorIs(t, test.expectedErr, err)
		})
	}
}

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})

	books := []entity.Book{
		repository.CreateBook("How to live in the beauty trash", SUCCESS),
		repository.CreateBook("How to not live in the beauty trash", SUCCESS),
		repository.CreateBook("Live in the beauty", FAILURE),
	}

	authorMock.EXPECT().GetAuthorBooks(gomock.Any(), gomock.Eq(SUCCESS)).Return([]entity.Book{books[0], books[1]}, nil)
	authorMock.EXPECT().GetAuthorBooks(gomock.Any(), gomock.Eq(FAILURE)).Return([]entity.Book{}, entity.ErrAuthorNotFound)

	tests := []struct {
		name        string
		target      AuthorUseCase
		authorID    string
		expected    []entity.Book
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			authorID:    SUCCESS,
			expected:    []entity.Book{books[0], books[1]},
			expectedErr: nil,
		},
		{
			name:        "failure case",
			target:      target,
			authorID:    FAILURE,
			expected:    []entity.Book{},
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := test.target.GetAuthorBooks(t.Context(), test.authorID)
			require.ErrorIs(t, test.expectedErr, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})

	successAuthor := repository.CreateAuthor(SUCCESS)

	authorMock.EXPECT().GetAuthorInfo(gomock.Any(), gomock.Eq(SUCCESS)).Return(successAuthor, nil)
	authorMock.EXPECT().GetAuthorInfo(gomock.Any(), gomock.Eq(FAILURE)).Return(entity.Author{}, entity.ErrAuthorNotFound)

	tests := []struct {
		name        string
		target      AuthorUseCase
		authorID    string
		expected    entity.Author
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			authorID:    SUCCESS,
			expected:    successAuthor,
			expectedErr: nil,
		},
		{
			name:        "failure case",
			target:      target,
			authorID:    FAILURE,
			expected:    entity.Author{},
			expectedErr: entity.ErrAuthorNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := test.target.GetAuthorInfo(t.Context(), test.authorID)
			require.ErrorIs(t, test.expectedErr, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestRegisterBook(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})

	successBook := repository.CreateBook(SUCCESS, SUCCESS)
	failureBook := repository.CreateBook(FAILURE, FAILURE)

	bookMock.EXPECT().CreateBook(gomock.Any(), gomock.Cond(func(x entity.Book) bool {
		return x.Name == SUCCESS
	})).Return(successBook, nil).AnyTimes()
	bookMock.EXPECT().CreateBook(gomock.Any(), gomock.Cond(func(x entity.Book) bool {
		return x.Name == FAILURE
	})).Return(entity.Book{}, entity.ErrBookAlreadyExists).AnyTimes()
	bookMock.EXPECT().CreateBook(gomock.Any(), gomock.Cond(func(x entity.Book) bool {
		return x.Name == FAILURE+"_1"
	})).Return(failureBook, nil).AnyTimes()
	outboxMock.EXPECT().SendMessage(gomock.Any(), repository.OutboxKindBook.String()+"_"+failureBook.ID, repository.OutboxKindBook, gomock.Any()).Return(entity.ErrBookNotFound).AnyTimes()
	outboxMock.EXPECT().SendMessage(gomock.Any(), gomock.Any(), repository.OutboxKindBook, gomock.Any()).Return(nil).AnyTimes()

	tests := []struct {
		name        string
		target      BooksUseCase
		bookID      string
		authorIDs   []string
		expected    entity.Book
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			bookID:      SUCCESS,
			authorIDs:   []string{SUCCESS},
			expected:    successBook,
			expectedErr: nil,
		},
		{
			name:        "failure case",
			target:      target,
			bookID:      FAILURE,
			authorIDs:   []string{},
			expected:    entity.Book{},
			expectedErr: entity.ErrBookAlreadyExists,
		},
		{
			name:        "failure send message",
			target:      target,
			bookID:      FAILURE + "_1",
			expected:    entity.Book{},
			expectedErr: entity.ErrBookNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := test.target.RegisterBook(t.Context(), test.bookID, test.authorIDs)
			require.ErrorIs(t, test.expectedErr, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})

	bookMock.EXPECT().UpdateBook(gomock.Any(), gomock.Cond(func(x entity.Book) bool {
		return x.Name == SUCCESS
	})).Return(nil)
	bookMock.EXPECT().UpdateBook(gomock.Any(), gomock.Cond(func(x entity.Book) bool {
		return x.Name == FAILURE
	})).Return(entity.ErrBookAlreadyExists)

	tests := []struct {
		name        string
		target      BooksUseCase
		bookID      string
		bookName    string
		authorIDs   []string
		expected    entity.Author
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			bookID:      SUCCESS,
			bookName:    SUCCESS,
			authorIDs:   []string{},
			expectedErr: nil,
		},
		{
			name:        "failure case",
			target:      target,
			bookID:      FAILURE,
			bookName:    FAILURE,
			authorIDs:   []string{},
			expectedErr: entity.ErrBookAlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.target.UpdateBook(t.Context(), test.bookID, test.bookName, test.authorIDs)
			require.ErrorIs(t, test.expectedErr, err)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorRepository(control)
	bookMock := mocks.NewMockBooksRepository(control)
	outboxMock := mocks.NewMockOutboxRepository(control)

	target := New(zaptest.NewLogger(t), authorMock, bookMock, outboxMock, &DumbTransactorImpl{})
	successBook := repository.CreateBook(SUCCESS)

	bookMock.EXPECT().GetBook(gomock.Any(), gomock.Eq(SUCCESS)).Return(successBook, nil)

	tests := []struct {
		name        string
		target      BooksUseCase
		bookID      string
		expected    entity.Book
		expectedErr error
	}{
		{
			name:        "success case",
			target:      target,
			bookID:      SUCCESS,
			expected:    successBook,
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := test.target.GetBook(t.Context(), test.bookID)
			require.ErrorIs(t, test.expectedErr, err)
			require.Equal(t, test.expected, result)
		})
	}
}
