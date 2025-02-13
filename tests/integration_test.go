package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dhaifley/apigo"
)

func TestMain(m *testing.M) {
	for _, arg := range os.Args {
		if arg == "-test.short=true" {
			// Skipping integration tests.
			os.Exit(0)
		}
	}

	su := os.Getenv("SUPERUSER")
	if su == "" {
		su = "admin"
	}

	sp := os.Getenv("SUPERUSER_PASSWORD")
	if sp == "" {
		sp = "admin"
	}

	os.Setenv("SUPERUSER", su)
	os.Setenv("SUPERUSER_PASSWORD", sp)

	svc := apigo.New()

	ctx := context.Background()

	go func(ctx context.Context) {
		if err := svc.Migrate(ctx); err != nil {
			fmt.Println("migrations error", err)

			os.Exit(1)
		}

		if err := svc.Start(ctx); err != nil {
			fmt.Println("server error", err)

			os.Exit(1)
		}
	}(ctx)

	time.Sleep(time.Second)

	code := m.Run()

	svc.Close(ctx)

	os.Exit(code)
}
