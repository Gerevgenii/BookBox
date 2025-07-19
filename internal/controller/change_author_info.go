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

func (i *implementation) ChangeAuthorInfo(ctx context.Context, req *library.ChangeAuthorInfoRequest) (ans *library.ChangeAuthorInfoResponse, erro error) {
	with, err := durations.GetMetricWithLabelValues("ChangeAuthorInfo")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("ChangeAuthorInfo called", traceID, zap.String("authorID", req.GetId()))
	start := time.Now()

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		if erro != nil {
			i.logger.Error("ChangeAuthorInfo error", zap.Error(erro), traceID)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("ChangeAuthorInfo completed", traceID)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = i.authorUseCase.UpdateAuthor(ctx, req.GetId(), req.GetName())

	if err != nil {
		return nil, i.convertErr(err)
	}

	return &library.ChangeAuthorInfoResponse{}, nil
}
