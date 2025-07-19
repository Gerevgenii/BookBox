package outbox

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/project/library/config"
	"github.com/project/library/generated/mocks"
	"github.com/project/library/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	t.Parallel()

	transactor := mocks.NewMockTransactor(gomock.NewController(t))
	outboxRepository := mocks.NewMockOutboxRepository(gomock.NewController(t))

	outbox := New(zaptest.NewLogger(t), outboxRepository, func(kind repository.OutboxKind) (KindHandler, error) {
		return nil, errors.New("unexpected error")
	}, nil, transactor)

	assert.NotNil(t, outbox)
}

type MyTransactor struct {
}

func (*MyTransactor) WithTx(ctx context.Context, function func(ctx context.Context) error) error {
	return function(ctx)
}

func TestStart(t *testing.T) {
	t.Parallel()

	transactor := &MyTransactor{}
	outboxRepository := mocks.NewMockOutboxRepository(gomock.NewController(t))

	outboxRepository.EXPECT().GetMessages(gomock.Any(), 2, gomock.Any()).Return([]repository.OutboxData{
		{
			IdempotencyKey: "121",
			Kind:           repository.OutboxKindBook,
		},
	}, nil).AnyTimes()
	outboxRepository.EXPECT().GetMessages(gomock.Any(), 3, gomock.Any()).Return([]repository.OutboxData{}, errors.New("unexpected error")).AnyTimes()
	outboxRepository.EXPECT().MarkAsProcessed(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	outbox := New(zaptest.NewLogger(t), outboxRepository, func(kind repository.OutboxKind) (KindHandler, error) {
		return nil, errors.New("unexpected error")
	}, &config.Config{
		Outbox: struct {
			Enabled         bool          `env:"OUTBOX_ENABLED"`
			Workers         int           `env:"OUTBOX_WORKERS"`
			BatchSize       int           `env:"OUTBOX_BATCH_SIZE"`
			WaitTimeMS      time.Duration `env:"OUTBOX_WAIT_TIME_MS"`
			InProgressTTLMS time.Duration `env:"OUTBOX_IN_PROGRESS_TTL_MS"`
			BookSendURL     string        `env:"OUTBOX_BOOK_SEND_URL"`
			AuthorSendURL   string        `env:"OUTBOX_AUTHOR_SEND_URL"`
		}{Enabled: true},
	}, transactor)
	assert.NotNil(t, outbox)

	ctx, cancelFunc := context.WithCancel(t.Context())
	wg := outbox.Start(ctx, 2, 2, time.Duration(2), time.Duration(2))
	time.Sleep(2 * time.Second)
	cancelFunc()
	wg.Wait()

	secondCtx, secondCancelFunc := context.WithCancel(t.Context())
	secondWg := outbox.Start(secondCtx, 2, 3, time.Duration(2), time.Duration(2))
	time.Sleep(2 * time.Second)
	secondCancelFunc()
	secondWg.Wait()
}
