package controller

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/generated/mocks"
	"github.com/project/library/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const FAILURE = "failure"
const SUCCESS = "success"

func TestAddBook(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorUseCase(control)
	bookMock := mocks.NewMockBooksUseCase(control)

	target := New(zaptest.NewLogger(t), bookMock, authorMock)

	successBook := entity.Book{
		ID:        SUCCESS,
		Name:      SUCCESS,
		AuthorIDs: nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bookMock.EXPECT().RegisterBook(gomock.Any(), FAILURE, gomock.Any()).Return(entity.Book{}, entity.ErrBookAlreadyExists)
	bookMock.EXPECT().RegisterBook(gomock.Any(), SUCCESS, gomock.Any()).Return(successBook, nil)

	tests := []struct {
		name         string
		target       *implementation
		req          *library.AddBookRequest
		expectedBook *library.AddBookResponse
		expectedErr  codes.Code
	}{
		{
			name:   FAILURE,
			target: target,
			req: &library.AddBookRequest{
				Name:      FAILURE,
				AuthorIds: nil,
			},
			expectedBook: nil,
			expectedErr:  codes.Internal,
		},
		{
			name:   FAILURE + "_uuid_format",
			target: target,
			req: &library.AddBookRequest{
				Name:      FAILURE,
				AuthorIds: []string{""},
			},
			expectedBook: nil,
			expectedErr:  codes.InvalidArgument,
		},
		{
			name:   SUCCESS,
			target: target,
			req: &library.AddBookRequest{
				Name:      SUCCESS,
				AuthorIds: nil,
			},
			expectedBook: &library.AddBookResponse{Book: &library.Book{
				Id:        successBook.ID,
				Name:      successBook.Name,
				AuthorId:  successBook.AuthorIDs,
				CreatedAt: timestamppb.New(successBook.CreatedAt),
				UpdatedAt: timestamppb.New(successBook.UpdatedAt),
			}},
			expectedErr: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualBook, err := test.target.AddBook(t.Context(), test.req)
			require.Equal(t, test.expectedBook, actualBook)
			s, ok := status.FromError(err)
			require.True(t, ok)
			if s != nil {
				require.Equal(t, test.expectedErr, s.Code())
			}
		})
	}
}

func TestChangeAuthorInfo(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorUseCase(control)
	bookMock := mocks.NewMockBooksUseCase(control)

	target := New(zaptest.NewLogger(t), bookMock, authorMock)

	authorMock.EXPECT().UpdateAuthor(gomock.Any(), gomock.Any(), FAILURE).Return(entity.ErrAuthorNotFound)
	authorMock.EXPECT().UpdateAuthor(gomock.Any(), gomock.Any(), SUCCESS).Return(nil)

	tests := []struct {
		name           string
		target         *implementation
		req            *library.ChangeAuthorInfoRequest
		expectedErr    codes.Code
		expectedAuthor *library.ChangeAuthorInfoResponse
	}{
		{
			name:   "invalid name",
			target: target,
			req: &library.ChangeAuthorInfoRequest{
				Name: "-_-",
				Id:   "",
			},
			expectedAuthor: nil,
			expectedErr:    codes.InvalidArgument,
		},
		{
			name:   "register error",
			target: target,
			req: &library.ChangeAuthorInfoRequest{
				Name: FAILURE,
				Id:   uuid.New().String(),
			},
			expectedAuthor: nil,
			expectedErr:    codes.NotFound,
		},
		{
			name:   SUCCESS,
			target: target,
			req: &library.ChangeAuthorInfoRequest{
				Name: SUCCESS,
				Id:   uuid.New().String(),
			},
			expectedAuthor: &library.ChangeAuthorInfoResponse{},
			expectedErr:    codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualAuthor, err := test.target.ChangeAuthorInfo(t.Context(), test.req)
			require.Equal(t, test.expectedAuthor, actualAuthor)
			s, ok := status.FromError(err)
			require.True(t, ok)
			if s != nil {
				require.Equal(t, test.expectedErr, s.Code())
			}
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorUseCase(control)
	bookMock := mocks.NewMockBooksUseCase(control)

	target := New(zaptest.NewLogger(t), bookMock, authorMock)

	failureID := uuid.New().String()
	successID := uuid.New().String()

	authorMock.EXPECT().GetAuthorInfo(gomock.Any(), failureID).Return(entity.Author{}, entity.ErrAuthorNotFound)
	authorMock.EXPECT().GetAuthorInfo(gomock.Any(), successID).Return(entity.Author{
		ID:   successID,
		Name: SUCCESS,
	}, nil)

	tests := []struct {
		name           string
		target         *implementation
		req            *library.GetAuthorInfoRequest
		expectedErr    codes.Code
		expectedAuthor *library.GetAuthorInfoResponse
	}{
		{
			name:   "invalid name",
			target: target,
			req: &library.GetAuthorInfoRequest{
				Id: "invalid id",
			},
			expectedAuthor: nil,
			expectedErr:    codes.InvalidArgument,
		},
		{
			name:   "register error",
			target: target,
			req: &library.GetAuthorInfoRequest{
				Id: failureID,
			},
			expectedAuthor: nil,
			expectedErr:    codes.NotFound,
		},
		{
			name:   SUCCESS,
			target: target,
			req: &library.GetAuthorInfoRequest{
				Id: successID,
			},
			expectedAuthor: &library.GetAuthorInfoResponse{
				Id:   successID,
				Name: SUCCESS,
			},
			expectedErr: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualAuthor, err := test.target.GetAuthorInfo(t.Context(), test.req)
			require.Equal(t, test.expectedAuthor, actualAuthor)
			s, ok := status.FromError(err)
			require.True(t, ok)
			if s != nil {
				require.Equal(t, test.expectedErr, s.Code())
			}
		})
	}
}

func TestGetBookInfo(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorUseCase(control)
	bookMock := mocks.NewMockBooksUseCase(control)

	target := New(zaptest.NewLogger(t), bookMock, authorMock)

	failureID := uuid.New().String()
	successID := uuid.New().String()

	bookMock.EXPECT().GetBook(gomock.Any(), failureID).Return(entity.Book{}, entity.ErrBookNotFound)
	successBook := entity.Book{
		ID:        successID,
		Name:      SUCCESS,
		AuthorIDs: nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	bookMock.EXPECT().GetBook(gomock.Any(), successID).Return(successBook, nil)

	tests := []struct {
		name         string
		target       *implementation
		req          *library.GetBookInfoRequest
		expectedErr  codes.Code
		expectedBook *library.GetBookInfoResponse
	}{
		{
			name:   "invalid name",
			target: target,
			req: &library.GetBookInfoRequest{
				Id: "invalid id",
			},
			expectedBook: nil,
			expectedErr:  codes.InvalidArgument,
		},
		{
			name:   "register error",
			target: target,
			req: &library.GetBookInfoRequest{
				Id: failureID,
			},
			expectedBook: nil,
			expectedErr:  codes.NotFound,
		},
		{
			name:   SUCCESS,
			target: target,
			req: &library.GetBookInfoRequest{
				Id: successID,
			},
			expectedBook: &library.GetBookInfoResponse{
				Book: &library.Book{
					Id:        successBook.ID,
					Name:      successBook.Name,
					AuthorId:  successBook.AuthorIDs,
					CreatedAt: timestamppb.New(successBook.CreatedAt),
					UpdatedAt: timestamppb.New(successBook.CreatedAt),
				},
			},
			expectedErr: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualBook, err := test.target.GetBookInfo(t.Context(), test.req)
			require.Equal(t, test.expectedBook, actualBook)
			s, ok := status.FromError(err)
			require.True(t, ok)
			if s != nil {
				require.Equal(t, test.expectedErr, s.Code())
			}
		})
	}
}

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorUseCase(control)
	bookMock := mocks.NewMockBooksUseCase(control)

	target := New(zaptest.NewLogger(t), bookMock, authorMock)

	successID := uuid.New().String()

	authorMock.EXPECT().RegisterAuthor(gomock.Any(), FAILURE).Return(entity.Author{}, entity.ErrAuthorAlreadyExists)
	successAuthor := entity.Author{
		Name: SUCCESS,
		ID:   successID,
	}
	authorMock.EXPECT().RegisterAuthor(gomock.Any(), SUCCESS).Return(successAuthor, nil)

	tests := []struct {
		name           string
		target         *implementation
		req            *library.RegisterAuthorRequest
		expectedErr    codes.Code
		expectedAuthor *library.RegisterAuthorResponse
	}{
		{
			name:   "invalid name",
			target: target,
			req: &library.RegisterAuthorRequest{
				Name: "",
			},
			expectedAuthor: nil,
			expectedErr:    codes.InvalidArgument,
		},
		{
			name:   "register error",
			target: target,
			req: &library.RegisterAuthorRequest{
				Name: FAILURE,
			},
			expectedAuthor: nil,
			expectedErr:    codes.Internal,
		},
		{
			name:   SUCCESS,
			target: target,
			req: &library.RegisterAuthorRequest{
				Name: SUCCESS,
			},
			expectedAuthor: &library.RegisterAuthorResponse{
				Id: successAuthor.ID,
			},
			expectedErr: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualAuthor, err := test.target.RegisterAuthor(t.Context(), test.req)
			require.Equal(t, test.expectedAuthor, actualAuthor)
			s, ok := status.FromError(err)
			require.True(t, ok)
			if s != nil {
				require.Equal(t, test.expectedErr, s.Code())
			}
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()

	control := gomock.NewController(t)
	authorMock := mocks.NewMockAuthorUseCase(control)
	bookMock := mocks.NewMockBooksUseCase(control)

	target := New(zaptest.NewLogger(t), bookMock, authorMock)

	bookMock.EXPECT().UpdateBook(gomock.Any(), gomock.Any(), FAILURE, gomock.Any()).Return(entity.ErrBookNotFound)
	bookMock.EXPECT().UpdateBook(gomock.Any(), gomock.Any(), SUCCESS, gomock.Any()).Return(nil)

	tests := []struct {
		name         string
		target       *implementation
		req          *library.UpdateBookRequest
		expectedErr  codes.Code
		expectedBook *library.UpdateBookResponse
	}{
		{
			name:   "invalid name",
			target: target,
			req: &library.UpdateBookRequest{
				Id:        "",
				Name:      "-_-",
				AuthorIds: []string{},
			},
			expectedBook: nil,
			expectedErr:  codes.InvalidArgument,
		},
		{
			name:   "register error",
			target: target,
			req: &library.UpdateBookRequest{
				Id:        uuid.New().String(),
				Name:      FAILURE,
				AuthorIds: []string{},
			},
			expectedBook: nil,
			expectedErr:  codes.NotFound,
		},
		{
			name:   SUCCESS,
			target: target,
			req: &library.UpdateBookRequest{
				Id:        uuid.New().String(),
				Name:      SUCCESS,
				AuthorIds: []string{},
			},
			expectedBook: &library.UpdateBookResponse{},
			expectedErr:  codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actualBook, err := test.target.UpdateBook(t.Context(), test.req)
			require.Equal(t, test.expectedBook, actualBook)
			s, ok := status.FromError(err)
			require.True(t, ok)
			if s != nil {
				require.Equal(t, test.expectedErr, s.Code())
			}
		})
	}
}
