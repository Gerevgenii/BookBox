package controller

import (
	"github.com/pkg/errors"
	"github.com/project/library/internal/entity"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) convertErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, entity.ErrAuthorNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, entity.ErrBookNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

var durations = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "library_durations_ms",
		Help:    "Durations in ms",
		Buckets: prometheus.DefBuckets,
	},
	[]string{
		"lever",
	})

func init() {
	prometheus.MustRegister(durations)
}
