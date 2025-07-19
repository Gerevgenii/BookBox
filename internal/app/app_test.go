package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jfrog/go-mockhttp"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

const (
	authorURLTest = "/author"
	bookURLTest   = "/book"
	SUCCESS       = "success"
)

func TestGlobalOutboxHandler(t *testing.T) {
	t.Parallel()
	successAuthorClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(authorURLTest)).Respond(mockhttp.Response())).HttpClient()

	successBookClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(bookURLTest)).Respond(mockhttp.Response())).HttpClient()

	tests := []struct {
		name              string
		client            *http.Client
		kind              repository.OutboxKind
		globalExpectedErr error
	}{
		{
			name:              "success author global handler",
			client:            successAuthorClient,
			kind:              repository.OutboxKindAuthor,
			globalExpectedErr: nil,
		},
		{
			name:              "success book global handler",
			client:            successBookClient,
			kind:              repository.OutboxKindBook,
			globalExpectedErr: nil,
		},
		{
			name:              "failure global handler",
			client:            successBookClient,
			kind:              repository.OutboxKindUndefined,
			globalExpectedErr: fmt.Errorf("unsupported outbox kind: %d", 0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			globalOutboxHand := globalOutboxHandler(test.client, bookURLTest, authorURLTest)
			require.NotNil(t, globalOutboxHand)
			_, err := globalOutboxHand(test.kind)
			require.Equal(t, err, test.globalExpectedErr)
		})
	}
}

func TestAuthorOutboxHandler(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")
	successAuthorClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(authorURLTest)).Respond(mockhttp.Response())).HttpClient()
	failureAuthorResponseClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(authorURLTest)).Respond(mockhttp.Response().StatusCode(http.StatusNotFound))).HttpClient()
	failureAuthorClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(authorURLTest)).ReturnError(err)).HttpClient()

	author := entity.Author{
		ID:        uuid.New().String(),
		Name:      SUCCESS,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	successData, err := json.Marshal(&author)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name              string
		client            *http.Client
		data              []byte
		globalExpectedErr string
	}{
		{
			name:              "success author global handler",
			client:            successAuthorClient,
			data:              successData,
			globalExpectedErr: "",
		},
		{
			name:              "failure unmarshal test",
			client:            successAuthorClient,
			data:              []byte{'a', 'b', 'o', 'b', 'a'},
			globalExpectedErr: "can not deserialize data in author outbox handler:",
		},
		{
			name:              "failure client test",
			client:            failureAuthorClient,
			data:              successData,
			globalExpectedErr: "Post",
		},
		{
			name:              "failure http response test",
			client:            failureAuthorResponseClient,
			data:              successData,
			globalExpectedErr: "failure code:",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			outboxHand := authorOutboxHandler(test.client, authorURLTest)
			require.NotNil(t, outboxHand)
			err := outboxHand(t.Context(), test.data)
			if test.globalExpectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), test.globalExpectedErr)
			}
		})
	}
}

func TestBookOutboxHandler(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")
	successBookClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(bookURLTest)).Respond(mockhttp.Response())).HttpClient()
	failureBookResponseClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(bookURLTest)).Respond(mockhttp.Response().StatusCode(http.StatusNotFound))).HttpClient()
	failureBookClient := mockhttp.NewClient(mockhttp.NewClientEndpoint().When(mockhttp.Request().POST(bookURLTest)).ReturnError(err)).HttpClient()

	book := entity.Book{
		ID:        uuid.New().String(),
		Name:      SUCCESS,
		AuthorIDs: []string{uuid.New().String()},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	successData, err := json.Marshal(&book)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name              string
		client            *http.Client
		data              []byte
		globalExpectedErr string
	}{
		{
			name:              "success author global handler",
			client:            successBookClient,
			data:              successData,
			globalExpectedErr: "",
		},
		{
			name:              "failure unmarshal test",
			client:            successBookClient,
			data:              []byte{'a', 'b', 'o', 'b', 'a'},
			globalExpectedErr: "can not deserialize data in book outbox handler:",
		},
		{
			name:              "failure client test",
			client:            failureBookClient,
			data:              successData,
			globalExpectedErr: "Post",
		},
		{
			name:              "failure http response test",
			client:            failureBookResponseClient,
			data:              successData,
			globalExpectedErr: "failure code:",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			outboxHand := bookOutboxHandler(test.client, bookURLTest)
			require.NotNil(t, outboxHand)
			err := outboxHand(t.Context(), test.data)
			if test.globalExpectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), test.globalExpectedErr)
			}
		})
	}
}
