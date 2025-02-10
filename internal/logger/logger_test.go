package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"testing"
	"testing/slogtest"

	"github.com/dhaifley/apigo/internal/logger"
)

func mockContext() context.Context {
	return context.WithValue(context.Background(), 5,
		"11223344-5566-7788-9900-aabbccddeeff")
}

func TestLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	h := logger.NewLogHandler(slog.NewJSONHandler(&buf, nil))

	results := func() []map[string]any {
		var ms []map[string]any

		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}

			var m map[string]any

			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}

			ms = append(ms, m)
		}

		return ms
	}

	err := slogtest.TestHandler(h, results)
	if err != nil {
		log.Fatal(err)
	}
}
