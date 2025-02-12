package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HealthHandler returns a route handler for /health requests.
func (s *Server) HealthHandler() http.Handler {
	r := chi.NewRouter()

	r.With(s.Stat, s.Trace).Get("/", s.GetHealthCheck)
	r.With(s.Stat, s.Trace, s.Auth).Post("/", s.PutHealthCheck)
	r.With(s.Stat, s.Trace, s.Auth).Patch("/", s.PutHealthCheck)
	r.With(s.Stat, s.Trace, s.Auth).Put("/", s.PutHealthCheck)

	return r
}

// HealthCheck values represent return information from health checks.
type HealthCheck struct {
	Service   string `json:"service,omitempty"`
	Version   string `json:"version,omitempty"`
	CommitID  string `json:"commit_id,omitempty"`
	BuildTime string `json:"build_time,omitempty"`
	Health    uint32 `json:"health,omitempty"`
}

// GetHealthCheck is the handler function for the health check path.
func (s *Server) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	res := &HealthCheck{
		Service: s.cfg.ServiceName(),
		Health:  s.Health(),
		Version: Version,
	}

	w.WriteHeader(int(res.Health))

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PutHealthCheck is the handler function for setting the server health code.
func (s *Server) PutHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeSuperUser); err != nil {
		s.error(err, w, r)

		return
	}

	req := &HealthCheck{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	s.SetHealth(req.Health)

	res := &HealthCheck{
		Service: s.cfg.ServiceName(),
		Health:  s.Health(),
	}

	w.WriteHeader(int(res.Health))

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// UpdateMetrics is used to periodically update the service metrics.
func (s *Server) UpdateMetrics(ctx context.Context,
) error {
	if s.metric == nil {
		return errors.New(errors.ErrUnavailable,
			"metrics not available for this server")
	}

	ctx, cancel := context.WithCancel(ctx)

	s.addCancelFunc(cancel)

	interval := time.Duration(0)

	go func(ctx context.Context) {
		for {
			tick := time.NewTimer(interval)

			interval = s.cfg.MetricInterval()

			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				ms := &runtime.MemStats{}

				runtime.ReadMemStats(ms)

				s.metric.Set(ctx, "alloc", int64(ms.Alloc))
				s.metric.Set(ctx, "total_alloc", int64(ms.TotalAlloc))
				s.metric.Set(ctx, "goroutines", int64(runtime.NumGoroutine()))

				if s.db != nil {
					dbStat := s.db.Stat()

					s.metric.Set(ctx, "db_open", int64(dbStat.TotalConns()))
					s.metric.Set(ctx, "db_in_use", int64(dbStat.AcquiredConns()))
					s.metric.Set(ctx, "db_idle", int64(dbStat.IdleConns()))
					s.metric.RecordDuration(ctx, "db_wait",
						dbStat.AcquireDuration())
				}
			}
		}
	}(ctx)

	return nil
}

// Trace wraps an http handler to include tracing information.
func (s *Server) Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tID := ""

		if s.tracer != nil {
			tc := propagation.TraceContext{}

			ctx = tc.Extract(ctx, propagation.HeaderCarrier(r.Header))

			peer := r.RemoteAddr

			remote := r.Header.Get("X-Forwarded-For")
			if remote == "" {
				remote = peer
			}

			route := chi.RouteContext(ctx).RoutePattern()

			scheme := "https"

			if strings.Contains(r.Host, "localhost") {
				scheme = "http"
			}

			ctx, span := s.tracer.Start(ctx, r.Method+" "+route,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.route", route),
					attribute.String("http.method", r.Method),
					attribute.String("http.target", r.URL.Path),
					attribute.String("http.flavor", fmt.Sprintf("%d.%d",
						r.ProtoMajor, r.ProtoMinor)),
					attribute.String("net.host.name", r.Host),
					attribute.String("http.scheme", scheme),
					attribute.String("http.client_ip", remote),
					attribute.String("net.sock.peer.addr", peer),
					attribute.Int("http.status_code", http.StatusOK),
				),
			)

			defer func() {
				if span != nil {
					span.End()
				}
			}()

			// Ensure the request and context contains tracing information.
			tc.Inject(ctx, propagation.HeaderCarrier(r.Header))

			tID = span.SpanContext().TraceID().String()
		}

		if tID == "" {
			if t, err := request.ContextTraceID(ctx); err == nil && t != "" {
				tID = t
			} else {
				if tu, err := uuid.NewRandom(); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to create UUID for trace_id",
						"error", err,
						"request", r)
				} else {
					tID = tu.String()
				}
			}
		}

		ctx = context.WithValue(ctx, request.CtxKeyTraceID, tID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Stat wraps an http handler to record server statistics.
func (s *Server) Stat(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mr := s.metric; mr != nil {
			ctx := r.Context()

			start := time.Now()

			operation := strings.ToLower(r.Method)

			route := chi.RouteContext(ctx).RoutePattern()

			defer func() {
				if mr != nil {
					mr.RecordDuration(ctx, "latency", time.Since(start),
						"route:"+route, "operation:"+operation)
				}
			}()

			mr.Increment(ctx, "requests",
				"route:"+route, "operation:"+operation)
		}

		next.ServeHTTP(w, r)
	})
}
