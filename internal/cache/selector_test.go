package cache_test

import (
	"net"
	"testing"

	"github.com/dhaifley/apid/internal/cache"
)

func TestServerList(t *testing.T) {
	t.Parallel()

	ss, err := cache.NewServerList("localhost:11211", "localhost:11211")
	if err != nil {
		t.Fatal(err)
	}

	count := 0

	err = ss.Each(func(net.Addr) error {
		count++

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Errorf("Expected count: 2, got: %v", count)
	}

	addr, err := ss.PickServer("test")
	if err != nil {
		t.Fatal(err)
	}

	exp := "127.0.0.1:11211"
	if addr.String() != exp {
		t.Errorf("Expected address: %v, got: %v", exp, addr.String())
	}
}
