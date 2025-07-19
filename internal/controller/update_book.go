package controller

import (
	"context"
	"time"

	"github.com/project/library/generated/api/library"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) UpdateBook(ctx context.Context, req *library.UpdateBookRequest) (ans *library.UpdateBookResponse, erro error) {
	with, err := durations.GetMetricWithLabelValues("UpdateBook")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("UpdateBook called", traceID, zap.String("bookID", req.GetId()))
	start := time.Now()

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		if erro != nil {
			i.logger.Error("UpdateBook error", zap.Error(erro), traceID)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("UpdateBook completed", traceID)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = i.booksUseCase.UpdateBook(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		return nil, i.convertErr(err)
	}

	return &library.UpdateBookResponse{}, nil
}
