// Package apigo provides an API service.
package apigo

import (
	"context"
	"net/http"
	"reflect"
	_ "time/tzdata"

	"github.com/dhaifley/apigo/db/migrations"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/metric"
	"github.com/dhaifley/apigo/internal/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"
)

// Service values are used to provide API services.
type Service struct {
	svr *server.Server
	mp  *sdkmetric.MeterProvider
	tp  *sdktrace.TracerProvider
	cfg *config.Config
	log logger.Logger
}

// New initializes a new service.
func New() *Service {
	svc := &Service{
		cfg: config.New("api"),
	}

	svc.cfg.Load(nil)

	svc.log = logger.New(svc.cfg.LogOut(), svc.cfg.LogFormat(),
		svc.cfg.LogLevel())

	return svc
}

// Handler returns the http handler function for the service.
func (s *Service) Version() string {
	return server.Version
}

// Handler returns the http handler function for the service.
func (s *Service) Handler() http.HandlerFunc {
	if s.svr == nil {
		return nil
	}

	return s.svr.Mux
}

// Start begins service operations.
func (s *Service) Start(ctx context.Context) error {
	var (
		err error
		mr  *metric.MetricRecorder
		tr  trace.Tracer
	)

	if s.cfg.MetricAddress() != "" {
		s.mp, err = newMeterProvider(ctx, s.cfg, s.log)
		if err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to create meter provider",
				"error", err)
		}

		mr = metric.NewRecorder(s.cfg, s.mp)
	}

	if s.cfg.TraceAddress() != "" {
		s.tp, err = newTracerProvider(ctx, s.cfg, s.log)
		if err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to create tracer provider",
				"error", err)
		} else {
			tr = s.tp.Tracer(s.cfg.ServiceName())
		}
	}

	s.svr, err = server.NewServer(s.cfg, s.log, mr, tr)
	if err != nil {
		return err
	}

	go func(ctx context.Context, svr *server.Server) {
		// Start emitting metrics.
		if err := svr.UpdateMetrics(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to emit server metrics",
				"error", err)
		}

		// Connect to the database.
		svr.ConnectSQL()

		// Get and update authentication configuration data.
		svr.UpdateAuthConfig()
	}(ctx, s.svr)

	return s.svr.Serve()
}

// Close shuts down service operations.
func (s *Service) Close(ctx context.Context) {
	s.svr.Shutdown(ctx)

	if s.mp != nil {
		if err := s.mp.Shutdown(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to shutdown meter provider",
				"error", err)
		}
	}

	if s.tp != nil {
		if err := s.tp.Shutdown(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to shutdown tracer provider",
				"error", err)
		}
	}
}

// Migrate will apply database migrations.
func (s *Service) Migrate(ctx context.Context) error {
	if err := migrations.Migrate(s.cfg, s.log); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to apply database migrations")
	}

	s.log.Log(ctx, logger.LvlInfo,
		"database migrations complete")

	return nil
}

type otlpErrorHandler struct {
	log logger.Logger
}

func (eh otlpErrorHandler) Handle(err error) {
	eh.log.Log(context.Background(), logger.LvlError,
		"telemetry error",
		"error", err)
}

// newTracerProvider initializes the tracer provider for the service.
func newTracerProvider(ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
) (*sdktrace.TracerProvider, error) {
	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	otel.SetErrorHandler(otlpErrorHandler{log: log})

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("",
			semconv.ServiceName(cfg.ServiceName()),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrTrace,
			"unable to create tracing resource for service")
	}

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.TraceAddress()),
		otlptracehttp.WithInsecure(),
	)

	var exp sdktrace.SpanExporter

	if exp, err = otlptrace.New(ctx, client); err != nil {
		return nil, errors.Wrap(err, errors.ErrTrace,
			"unable to create new otlp trace exporter",
			"address", cfg.TraceAddress())
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

// newMeterProvider initializes the meter provider for the service.
func newMeterProvider(ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
) (*sdkmetric.MeterProvider, error) {
	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	otel.SetErrorHandler(otlpErrorHandler{log: log})

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("",
			semconv.ServiceName(cfg.ServiceName()),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrMetric,
			"unable to create metrics resource for service")
	}

	exp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.MetricAddress()),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrMetric,
			"unable to create new metrics exporter",
			"address", cfg.MetricAddress())
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp,
			sdkmetric.WithInterval(cfg.MetricInterval()))),
		sdkmetric.WithResource(r),
	)

	otel.SetMeterProvider(mp)

	return mp, nil
}
