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

func (i *implementation) GetAuthorInfo(ctx context.Context, req *library.GetAuthorInfoRequest) (ans *library.GetAuthorInfoResponse, erro error) {
	with, err := durations.GetMetricWithLabelValues("GetAuthorInfo")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("GetAuthorInfo called", traceID, zap.String("authorID", req.GetId()))
	start := time.Now()

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		authorResponseId := zap.String("authorResponseID", ans.GetId())
		if erro != nil {
			i.logger.Error("GetAuthorInfo error", zap.Error(erro), traceID, authorResponseId)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("GetAuthorInfo completed", traceID, authorResponseId)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthorInfo(ctx, req.GetId())

	if err != nil {
		return nil, i.convertErr(err)
	}

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
