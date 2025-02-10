package config_test

import (
	"crypto/rsa"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/config"
)

func TestAuthConfig(t *testing.T) {
	t.Parallel()

	exp := "test"

	cfg := config.NewDefault()

	cfg.SetAuth(&config.AuthConfig{
		TokenHMACKey:          []byte(exp),
		TokenPrivateKey:       []byte(exp),
		TokenPublicKey:        []byte(exp),
		TokenJWKS:             "{}",
		TokenWellKnown:        exp,
		TokenExpiresIn:        time.Second * 10,
		TokenRefreshExpiresIn: time.Second * 10,
		TokenIssuer:           exp,
		UpdateInterval:        time.Second,
		IdentityDomain:        exp,
	})

	cfg.SetAuthTokenJWKS(map[string]*rsa.PublicKey{})

	if string(cfg.AuthTokenHMACKey()) != exp {
		t.Errorf("Expected HMAC Key: %v, got: %v",
			exp, string(cfg.AuthTokenHMACKey()))
	}

	if cfg.AuthTokenWellKnown() != exp {
		t.Errorf("Expected .wellknown: %v, got: %v",
			exp, cfg.AuthTokenWellKnown())
	}

	if cfg.AuthTokenJWKSPublicKey(exp) != nil {
		t.Errorf("Expected JWKS public key: null, got: %v",
			cfg.AuthTokenJWKSPublicKey(exp))
	}

	if cfg.AuthTokenJWKSLength() != 0 {
		t.Errorf("Expected jwks length: 0, got: %v", cfg.AuthTokenJWKSLength())
	}

	if cfg.AuthTokenExpiresIn() != 10*time.Second {
		t.Errorf("Expected expiration: 10s, got: %v", cfg.AuthTokenExpiresIn())
	}

	if cfg.AuthTokenRefreshExpiresIn() != 10*time.Second {
		t.Errorf("Expected refresh expiration: 10s, got: %v",
			cfg.AuthTokenRefreshExpiresIn())
	}

	if cfg.AuthTokenIssuer() != exp {
		t.Errorf("Expected file: %v, got: %v", exp, cfg.AuthTokenIssuer())
	}

	if cfg.AuthUpdateInterval() != time.Second {
		t.Errorf("Expected interval: 1s, got: %v", cfg.AuthUpdateInterval())
	}

	if cfg.AuthIdentityDomain() != exp {
		t.Errorf("Expected identity domain: %v, got: %v",
			exp, cfg.AuthIdentityDomain())
	}
}
