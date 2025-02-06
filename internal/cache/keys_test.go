package cache_test

import (
	"testing"

	"github.com/dhaifley/apid/internal/cache"
)

func TestCacheKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		exp string
		run func() string
	}{
		{
			exp: "Account::test",
			run: func() string { return cache.KeyAccount("test") },
		},
		{
			exp: "Account::Name::test",
			run: func() string { return cache.KeyAccountName("test") },
		},
		{
			exp: "User::test",
			run: func() string { return cache.KeyUser("test") },
		},
		{
			exp: "User::Details::test",
			run: func() string { return cache.KeyUserDetails("test") },
		},
		{
			exp: "Token::Auth::test",
			run: func() string { return cache.KeyAuthToken("test") },
		},
		{
			exp: "Token::test",
			run: func() string { return cache.KeyToken("test") },
		},
		{
			exp: "Resource::test",
			run: func() string { return cache.KeyResource("test") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.exp, func(t *testing.T) {
			t.Parallel()

			res := tt.run()

			if tt.exp != res {
				t.Errorf("Expected key: %v, got: %v", tt.exp, res)
			}
		})
	}
}
