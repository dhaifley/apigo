package metric_test

import (
	"context"
	"testing"
	"time"

	"github.com/dhaifley/apid/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	mp := sdkmetric.NewMeterProvider()

	r := metric.NewRecorder(nil, mp)

	ctx := context.Background()

	r.Set(ctx, "counter", 0)
	r.Add(ctx, "counter", 1)
	r.Increment(ctx, "counter")

	r.RecordValue(ctx, "histogram", 1.0, "test:test")
	r.RecordDuration(ctx, "histogram", time.Duration(1), "test:test")

	if r.Len() != 2 {
		t.Errorf("Expected recorder to contain 2 metrics, got: %v", r.Len())
	}
}
