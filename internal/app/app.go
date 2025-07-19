package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/outbox"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/db"

	grpcRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/project/library/config"
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/exporters/jaeger"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	shutdown := initTracer(logger, cfg.Observability.JaegerURL)
	defer func() {
		err := shutdown(ctx)

		if err != nil {
			logger.Error("can not shutdown jaeger collector", zap.Error(err))
		}
	}()

	go runMetricsServer(logger, cfg.Observability.MetricsPort)

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)
	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		return
	}
	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepository(logger, dbPool)
	outboxRepository := repository.NewOutbox(dbPool)

	transactor := repository.NewTransactor(dbPool)
	runOutbox(ctx, cfg, logger, outboxRepository, transactor)

	useCases := library.New(logger, repo, repo, outboxRepository, transactor)

	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)

	<-ctx.Done()
	const param = 3
	time.Sleep(time.Second * param)
}

func runMetricsServer(logger *zap.Logger, port string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Error("Metrics server error", zap.Error(err))
	}
}

func initTracer(l *zap.Logger, url string) func(context.Context) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

	if err != nil {
		l.Fatal("can not create jaeger collector", zap.Error(err))
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("library-service"),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown
}

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	const (
		timeoutConst               time.Duration = 30
		keepAliveConst             time.Duration = 180
		idleConnTimeoutConst       time.Duration = 90
		tlSHandshakeTimeoutConst   time.Duration = 15
		expectContinueTimeoutConst time.Duration = 2
		maxIdleConnsConst          int           = 100
	)
	dialer := &net.Dialer{
		Timeout:   timeoutConst * time.Second,
		KeepAlive: keepAliveConst * time.Second,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxIdleConnsConst,
		MaxConnsPerHost:       maxIdleConnsConst,
		IdleConnTimeout:       idleConnTimeoutConst * time.Second,
		TLSHandshakeTimeout:   tlSHandshakeTimeoutConst * time.Second,
		ExpectContinueTimeout: expectContinueTimeoutConst * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}

	client := new(http.Client)
	client.Transport = transport

	globalHandler := globalOutboxHandler(client, cfg.Outbox.BookSendURL, cfg.Outbox.AuthorSendURL)
	outboxService := outbox.New(logger, outboxRepository, globalHandler, cfg, transactor)

	outboxService.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.WaitTimeMS,
		cfg.Outbox.InProgressTTLMS,
	)
}

func globalOutboxHandler(
	client *http.Client,
	bookURL string,
	authorURL string,
) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(client, bookURL), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(client, authorURL), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func authorOutboxHandler(client *http.Client, url string) outbox.KindHandler {
	return func(_ context.Context, data []byte) (txErr error) {
		author := entity.Author{}
		err := json.Unmarshal(data, &author)

		if err != nil {
			return fmt.Errorf("can not deserialize data in author outbox handler: %w", err)
		}

		response, err := client.Post(url, "application/json", strings.NewReader(author.ID)) //nolint:bodyclose // Because I already do it
		if err != nil {
			return err
		}

		defer func(closer io.ReadCloser) {
			if err := closer.Close(); err != nil {
				txErr = fmt.Errorf("can not close author outbox: %w", err)
			}
		}(response.Body)

		const httpRequestNumber int = 2
		if response.StatusCode/100 != httpRequestNumber {
			return fmt.Errorf("failure code: %d", response.StatusCode)
		}
		return nil
	}
}

func bookOutboxHandler(client *http.Client, url string) outbox.KindHandler {
	return func(_ context.Context, data []byte) (txErr error) {
		book := entity.Book{}
		err := json.Unmarshal(data, &book)

		if err != nil {
			return fmt.Errorf("can not deserialize data in book outbox handler: %w", err)
		}

		response, err := client.Post(url, "application/json", strings.NewReader(book.ID)) //nolint:bodyclose // Because I already do it
		if err != nil {
			return err
		}

		defer func(closer io.ReadCloser) {
			if err := closer.Close(); err != nil {
				txErr = fmt.Errorf("can not close author outbox: %w", err)
			}
		}(response.Body)

		const httpRequestNumber int = 2
		if response.StatusCode/100 != httpRequestNumber {
			return fmt.Errorf("failure code: %d", response.StatusCode)
		}
		return nil
	}
}

func runRest(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	mux := grpcRuntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)

	if err != nil {
		logger.Error("can not register grpc gateway", zap.Error(err))
		os.Exit(-1)
	}

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	logger.Info("gateway listening at port", zap.String("port", gatewayPort))

	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(cfg *config.Config, logger *zap.Logger, libraryService generated.LibraryServer) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)

	if err != nil {
		logger.Error("can not open tcp socket", zap.Error(err))
		os.Exit(-1)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			),
		),
		grpc.StreamInterceptor(
			otelgrpc.StreamServerInterceptor(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			),
		),
	)
	reflection.Register(s)

	generated.RegisterLibraryServer(s, libraryService)

	logger.Info("grpc server listening at port", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}
