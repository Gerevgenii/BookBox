package controller

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) AddBook(ctx context.Context, req *library.AddBookRequest) (ans *library.AddBookResponse, erro error) {
	with, err := durations.GetMetricWithLabelValues("AddBook")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("AddBook called", traceID)
	start := time.Now()

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		if erro != nil {
			i.logger.Error("AddBook error", zap.Error(erro), traceID)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("AddBook completed", traceID)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.RegisterBook(ctx, req.GetName(), req.GetAuthorIds())

	if err != nil {
		return nil, i.convertErr(err)
	}

	return &library.AddBookResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
