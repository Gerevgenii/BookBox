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

func (i *implementation) RegisterAuthor(ctx context.Context, req *library.RegisterAuthorRequest) (ans *library.RegisterAuthorResponse, erro error) {
	with, err := durations.GetMetricWithLabelValues("RegisterAuthor")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("RegisterAuthor called", traceID)
	start := time.Now()

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		authorResponseId := zap.String("authorResponseID", ans.GetId())
		if erro != nil {
			i.logger.Error("RegisterAuthor error", zap.Error(erro), traceID, authorResponseId)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("RegisterAuthor completed", traceID, authorResponseId)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())

	if err != nil {
		return nil, i.convertErr(err)
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
