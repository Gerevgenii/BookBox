package controller

import (
	"context"
	"time"

	"github.com/project/library/generated/api/library"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *implementation) GetBookInfo(ctx context.Context, req *library.GetBookInfoRequest) (ans *library.GetBookInfoResponse, erro error) {
	with, err := durations.GetMetricWithLabelValues("GetBookInfo")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("GetBookInfo called", traceID, zap.String("authorID", req.GetId()))
	start := time.Now()

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		bookResponseId := zap.String("authorResponseID", ans.GetBook().GetId())
		if erro != nil {
			i.logger.Error("GetBookInfo error", zap.Error(erro), traceID, bookResponseId)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("GetBookInfo completed", traceID, bookResponseId)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBook(ctx, req.GetId())

	if err != nil {
		return nil, i.convertErr(err)
	}

	return &library.GetBookInfoResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.CreatedAt),
		},
	}, nil
}
