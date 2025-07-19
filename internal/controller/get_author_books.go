package controller

import (
	"time"

	"github.com/project/library/generated/api/library"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetAuthorBooks(req *library.GetAuthorBooksRequest, out library.Library_GetAuthorBooksServer) (erro error) {
	with, err := durations.GetMetricWithLabelValues("GetAuthorBooks")
	if err != nil {
		i.logger.Error("Can't get duration metric", zap.Error(err))
	}

	ctx := out.Context()
	var traceID = zap.String("traceID", trace.SpanFromContext(ctx).SpanContext().TraceID().String())
	i.logger.Info("GetAuthorBooks called", traceID, zap.String("authorID", req.GetAuthorId()))
	start := time.Now()

	// TODO: Create books log!

	defer func() {
		with.Observe(float64(time.Since(start).Milliseconds()))
		if erro != nil {
			i.logger.Error("GetAuthorBooks error", zap.Error(erro), traceID)
			trace.SpanFromContext(ctx).RecordError(erro)
		} else {
			i.logger.Info("GetAuthorBooks completed", traceID)
		}
		trace.SpanFromContext(ctx).End()
	}()

	if err := req.ValidateAll(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.authorUseCase.GetAuthorBooks(ctx, req.GetAuthorId())

	if err != nil {
		return i.convertErr(err)
	}

	for _, book := range books {
		errOfSending := out.Send(&library.Book{
			Id:       book.ID,
			Name:     book.Name,
			AuthorId: book.AuthorIDs,
		})
		if errOfSending != nil {
			return i.convertErr(errOfSending)
		}
	}

	return nil
}
