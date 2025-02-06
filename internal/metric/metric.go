// Package metric contains types and functions used for recording metrics
// telemetry data about the service operation.
package metric

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/dhaifley/apid/internal/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Recorder values are used to store and format metrics data.
type Recorder interface {
	Add(ctx context.Context, name string, value int64, tags ...string)
	Increment(ctx context.Context, name string, tags ...string)
	Set(ctx context.Context, name string, value int64, tags ...string)
	RecordDuration(ctx context.Context, name string, value time.Duration,
		tags ...string)
	RecordValue(ctx context.Context, name string, value float64, tags ...string)
}

// MetricRecorder values are used to store and format metrics data.
type MetricRecorder struct {
	sync.RWMutex
	cfg   *config.Config
	m     map[string]any
	meter metric.Meter
}

// NewRecorder initializes and returns a new metrics data recorder.
func NewRecorder(cfg *config.Config, mp metric.MeterProvider) *MetricRecorder {
	if cfg == nil {
		cfg = &config.Config{}
	}

	if mp == nil || (reflect.ValueOf(mp).Kind() == reflect.Ptr &&
		reflect.ValueOf(mp).IsNil()) {
		return nil
	}

	r := &MetricRecorder{
		cfg: cfg,
		m:   make(map[string]any),
		meter: mp.Meter(cfg.ServiceName(),
			metric.WithInstrumentationVersion(cfg.MetricVersion())),
	}

	return r
}

// Len returns the number of metrics recorded.
func (r *MetricRecorder) Len() int {
	r.RLock()
	defer r.RUnlock()

	return len(r.m)
}

// getAttributes is used to parse a set of tags into metric attributes.
func (r *MetricRecorder) getAttributes(tags []string,
) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("service", r.cfg.ServiceName()),
	}

	for _, t := range tags {
		ts := strings.SplitN(t, ":", 2)

		if len(ts) == 2 {
			attrs = append(attrs, attribute.String(ts[0], ts[1]))
		} else if len(ts) == 1 {
			attrs = append(attrs, attribute.String(ts[0], ""))
		}
	}

	return attrs
}

// Add increases a named counter metric by a value.
func (r *MetricRecorder) Add(ctx context.Context,
	name string,
	value int64,
	tags ...string,
) {
	r.Lock()
	defer r.Unlock()

	var err error

	m, ok := r.m[name]
	if !ok {
		m, err = r.meter.Int64Counter(name)
		if err != nil {
			return
		}

		r.m[name] = m
	}

	mv, ok := m.(metric.Int64Counter)
	if !ok {
		mv, err = r.meter.Int64Counter(name)
		if err != nil {
			return
		}

		r.m[name] = mv
	}

	mv.Add(ctx, int64(value), metric.WithAttributes(r.getAttributes(tags)...))
}

// Increment increases a named counter metric by one.
func (r *MetricRecorder) Increment(ctx context.Context,
	name string,
	tags ...string,
) {
	r.Lock()
	defer r.Unlock()

	var err error

	m, ok := r.m[name]
	if !ok {
		m, err = r.meter.Int64Counter(name)
		if err != nil {
			return
		}

		r.m[name] = m
	}

	mv, ok := m.(metric.Int64Counter)
	if !ok {
		mv, err = r.meter.Int64Counter(name)
		if err != nil {
			return
		}

		r.m[name] = mv
	}

	mv.Add(ctx, 1, metric.WithAttributes(r.getAttributes(tags)...))
}

// Set assigns a counter metric a specific value.
func (r *MetricRecorder) Set(ctx context.Context,
	name string,
	value int64,
	tags ...string,
) {
	r.Lock()
	defer r.Unlock()

	m, err := r.meter.Int64Counter(name)
	if err != nil {
		return
	}

	r.m[name] = m

	mv, ok := m.(metric.Int64Counter)
	if !ok {
		mv, err = r.meter.Int64Counter(name)
		if err != nil {
			return
		}

		r.m[name] = mv
	}

	mv.Add(ctx, value, metric.WithAttributes(r.getAttributes(tags)...))
}

// RecordDuration increments a duration bucket in a histogram metric by one.
func (r *MetricRecorder) RecordDuration(ctx context.Context,
	name string,
	value time.Duration,
	tags ...string,
) {
	r.Lock()
	defer r.Unlock()

	var err error

	m, ok := r.m[name]
	if !ok {
		m, err = r.meter.Float64Histogram(name)
		if err != nil {
			return
		}

		r.m[name] = m
	}

	mv, ok := m.(metric.Float64Histogram)
	if !ok {
		mv, err = r.meter.Float64Histogram(name)
		if err != nil {
			return
		}

		r.m[name] = mv
	}

	mv.Record(ctx, value.Seconds(),
		metric.WithAttributes(r.getAttributes(tags)...))
}

// RecordValue increments a value bucket in a histogram metric by one.
func (r *MetricRecorder) RecordValue(ctx context.Context,
	name string,
	value float64,
	tags ...string,
) {
	r.Lock()
	defer r.Unlock()

	var err error

	m, ok := r.m[name]
	if !ok {
		m, err = r.meter.Float64Histogram(name)
		if err != nil {
			return
		}

		r.m[name] = m
	}

	mv, ok := m.(metric.Float64Histogram)
	if !ok {
		mv, err = r.meter.Float64Histogram(name)
		if err != nil {
			return
		}

		r.m[name] = mv
	}

	mv.Record(ctx, value, metric.WithAttributes(r.getAttributes(tags)...))
}
