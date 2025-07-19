package outbox

import (
	"context"
	"sync"
	"time"

	"github.com/project/library/config"
	"github.com/project/library/internal/usecase/repository"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	outboxSuccessTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_success_total",
		Help: "Total number of successfully processed outbox messages",
	}, []string{"kind"})

	outboxFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_failed_total",
		Help: "Total number of failed outbox message processing attempts",
	}, []string{"kind"})

	outboxHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "library_durations_ms",
			Help:    "Durations in ms",
			Buckets: prometheus.DefBuckets,
		},
		[]string{
			"lever",
		})
)

func init() {
	prometheus.MustRegister(outboxSuccessTotal, outboxFailedTotal, outboxHistogram)
}

type GlobalHandler = func(kind repository.OutboxKind) (KindHandler, error)
type KindHandler = func(ctx context.Context, data []byte) error

type Outbox interface {
	Start(ctx context.Context, workers int, batchSize int, waitTime time.Duration, inProgressTTL time.Duration) *sync.WaitGroup
}

var _ Outbox = (*outboxImpl)(nil)

type outboxImpl struct {
	logger           *zap.Logger
	outboxRepository repository.OutboxRepository
	globalHandler    GlobalHandler
	cfg              *config.Config
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	globalHandler GlobalHandler,
	cfg *config.Config,
	transactor repository.Transactor,
) *outboxImpl {
	return &outboxImpl{
		logger:           logger,
		outboxRepository: outboxRepository,
		globalHandler:    globalHandler,
		cfg:              cfg,
		transactor:       transactor,
	}
}

func (o *outboxImpl) Start(
	ctx context.Context,
	workers int,
	batchSize int,
	waitTime time.Duration,
	inProgressTTL time.Duration,
) *sync.WaitGroup {
	wg := new(sync.WaitGroup)

	for workerID := 1; workerID <= workers; workerID++ {
		wg.Add(1)
		go o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)
	}
	return wg
}

func (o *outboxImpl) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	batchSize int,
	waitTIme time.Duration,
	inProgressTTL time.Duration,
) {
	defer wg.Done()
	for {
		time.Sleep(waitTIme)
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !o.cfg.Outbox.Enabled {
			continue
		}

		err := o.transactor.WithTx(ctx, func(ctx context.Context) error {
			messages, getMessageErr := o.outboxRepository.GetMessages(ctx, batchSize, inProgressTTL)

			if getMessageErr != nil {
				o.logger.Error("can not fetch messages from outbox", zap.Error(getMessageErr))
				return getMessageErr
			}

			// o.logger.Info("messages fetched", zap.Int("size", len(messages)))
			successKeys := make([]string, 0, len(messages))

			for i := 0; i < len(messages); i++ {

				message := messages[i]
				key := message.IdempotencyKey

				kindHandler, err := o.globalHandler(message.Kind)
				start := time.Now()

				metricOutboxHistogram, err := outboxHistogram.GetMetricWithLabelValues(message.Kind.String())
				if err != nil {
					o.logger.Error("Can't get latency metric", zap.Error(err))
				}

				if err != nil {
					o.logger.Error("unexpected kind", zap.Error(err))
					metricOutboxHistogram.Observe(float64(time.Since(start).Milliseconds()))
					continue
				}

				err = kindHandler(ctx, message.RawData)

				if err != nil {
					o.logger.Error("kind error", zap.Error(err))
					metricOutboxHistogram.Observe(float64(time.Since(start).Milliseconds()))
					continue
				}

				successKeys = append(successKeys, key)
				failedTotal, err := outboxFailedTotal.GetMetricWithLabelValues(message.Kind.String())
				if err != nil {
					o.logger.Error("Can't get failures metric counter", zap.Error(err))
				}
				failedTotal.Inc()
				successTotal, err := outboxSuccessTotal.GetMetricWithLabelValues(message.Kind.String())
				if err != nil {
					o.logger.Error("Can't get successes metric counter", zap.Error(err))
				}
				successTotal.Inc()
				metricOutboxHistogram.Observe(float64(time.Since(start).Milliseconds()))
			}

			getMessageErr = o.outboxRepository.MarkAsProcessed(ctx, successKeys)
			if getMessageErr != nil {
				o.logger.Error("mark as processed outbox error", zap.Error(getMessageErr))
				return getMessageErr
			}

			return nil
		})

		if err != nil {
			o.logger.Error("worker stage error", zap.Error(err))
		}
	}
}
