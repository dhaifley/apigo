package config

import (
	"bytes"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

const (
	KeyAuthTokenHMACKey          = "auth/token/hmac_key"
	KeyAuthTokenPrivateKey       = "auth/token/private_key"
	KeyAuthTokenPublicKey        = "auth/token/public_key"
	KeyAuthTokenPublicKeyFile    = "auth/token/public_key_file"
	KeyAuthTokenWellKnown        = "auth/token/well_known"
	KeyAuthTokenJWKS             = "auth/token/jwks"
	KeyAuthTokenExpiresIn        = "auth/token/expires_in"
	KeyAuthTokenRefreshExpiresIn = "refresh/token/expires_in"
	KeyAuthTokenIssuer           = "auth/token/issuer"
	KeyAuthUpdateInterval        = "auth/update_interval"
	KeyAuthIdentityDomain        = "auth/identity_domain"

	DefaultAuthTokenJWKS             = "{}"
	DefaultAuthTokenWellKnown        = ""
	DefaultAuthTokenExpiresIn        = time.Hour * 24
	DefaultAuthTokenRefreshExpiresIn = time.Hour * 24 * 30
	DefaultAuthTokenIssuer           = "api"
	DefaultAuthUpdateInterval        = time.Second * 30
	DefaultAuthIdentityDomain        = ""
)

// AuthConfig values represent authentication configuration data.
type AuthConfig struct {
	TokenHMACKey          []byte        `json:"token_hmac_key,omitempty"           yaml:"token_hmac_key,omitempty"`
	TokenPrivateKey       []byte        `json:"token_private_key,omitempty"        yaml:"token_private_key,omitempty"`
	TokenPublicKey        []byte        `json:"token_public_key,omitempty"         yaml:"token_public_key,omitempty"`
	TokenJWKS             string        `json:"token_jwks,omitempty"               yaml:"token_jwks,omitempty"`
	TokenWellKnown        string        `json:"token_well_known,omitempty"         yaml:"token_well_known,omitempty"`
	TokenExpiresIn        time.Duration `json:"token_expires_in,omitempty"         yaml:"token_expires_in,omitempty"`
	TokenRefreshExpiresIn time.Duration `json:"token_refresh_expires_in,omitempty" yaml:"token_refresh_expires_in,omitempty"`
	TokenIssuer           string        `json:"token_issuer,omitempty"             yaml:"token_issuer,omitempty"`
	UpdateInterval        time.Duration `json:"update_interval,omitempty"          yaml:"update_interval,omitempty"`
	IdentityDomain        string        `json:"identity_domain,omitempty"          yaml:"identity_domain,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *AuthConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyAuthTokenHMACKey)); v != "" {
		if _, err := hex.Decode(c.TokenHMACKey, []byte(v)); err != nil {
			c.TokenHMACKey = []byte{}
		}
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenPrivateKey)); v != "" {
		if _, err := hex.Decode(c.TokenPrivateKey, []byte(v)); err != nil {
			c.TokenPrivateKey = []byte{}
		}
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenPublicKey)); v != "" {
		if _, err := hex.Decode(c.TokenPublicKey, []byte(v)); err != nil {
			c.TokenPublicKey = []byte{}
		}
	}

	if len(c.TokenPublicKey) == 0 {
		if f := os.Getenv(ReplaceEnv(KeyAuthTokenPublicKeyFile)); f != "" {
			if pk, err := os.ReadFile(f); err == nil && len(pk) > 0 {
				c.TokenPublicKey = pk
			}
		}
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenJWKS)); v != "" {
		c.TokenJWKS = v
	}

	if c.TokenJWKS == "" {
		c.TokenJWKS = DefaultAuthTokenJWKS
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenWellKnown)); v != "" {
		c.TokenWellKnown = v
	}

	if c.TokenWellKnown == "" {
		c.TokenWellKnown = DefaultAuthTokenWellKnown
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenExpiresIn)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultAuthTokenExpiresIn
		}

		c.TokenExpiresIn = v
	}

	if c.TokenExpiresIn == 0 {
		c.TokenExpiresIn = DefaultAuthTokenExpiresIn
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenRefreshExpiresIn)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultAuthTokenRefreshExpiresIn
		}

		c.TokenRefreshExpiresIn = v
	}

	if c.TokenRefreshExpiresIn == 0 {
		c.TokenRefreshExpiresIn = DefaultAuthTokenRefreshExpiresIn
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthTokenIssuer)); v != "" {
		c.TokenIssuer = v
	}

	if c.TokenIssuer == "" {
		c.TokenIssuer = DefaultAuthTokenIssuer
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthUpdateInterval)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultAuthUpdateInterval
		}

		c.UpdateInterval = v
	}

	if c.UpdateInterval == 0 {
		c.UpdateInterval = DefaultAuthUpdateInterval
	}

	if v := os.Getenv(ReplaceEnv(KeyAuthIdentityDomain)); v != "" {
		c.IdentityDomain = v
	}

	if c.IdentityDomain == "" {
		c.IdentityDomain = DefaultAuthIdentityDomain
	}
}

// AuthTokenHMACKey returns the HMAC key used for token encryption.
func (c *Config) AuthTokenHMACKey() []byte {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return nil
	}

	return c.auth.TokenHMACKey
}

// AuthTokenPrivateKey returns the private key used for token encryption.
func (c *Config) AuthTokenPrivateKey() []byte {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return nil
	}

	return c.auth.TokenPrivateKey
}

// AuthTokenPublicKey returns the public key used for token encryption.
func (c *Config) AuthTokenPublicKey() []byte {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return nil
	}

	return c.auth.TokenPublicKey
}

// AuthTokenWellKnown returns the path used to retrieve auth well known data.
func (c *Config) AuthTokenWellKnown() string {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return DefaultAuthTokenWellKnown
	}

	return c.auth.TokenWellKnown
}

// AuthTokenJWKSPublicKey returns the public key for a specified key ID.
func (c *Config) AuthTokenJWKSPublicKey(kid string) *rsa.PublicKey {
	if c.auth == nil {
		return nil
	}

	jwks := c.auth.TokenJWKS
	if jwks == "" || jwks == "{}" {
		return nil
	}

	jm := map[string]*rsa.PublicKey{}

	if err := json.Unmarshal([]byte(jwks), &jm); err != nil {
		return nil
	}

	res, ok := jm[kid]
	if !ok || res == nil {
		return nil
	}

	return res
}

// AuthTokenJWKSLength returns number of JWKS keys retrieved.
func (c *Config) AuthTokenJWKSLength() int {
	if c.auth == nil {
		return 0
	}

	jwks := c.auth.TokenJWKS
	if jwks == "" || jwks == "{}" {
		return 0
	}

	jm := map[string]*rsa.PublicKey{}

	if err := json.Unmarshal([]byte(jwks), &jm); err != nil {
		return 0
	}

	return len(jm)
}

// AuthTokenExpiresIn returns the duration of time the token is valid.
func (c *Config) AuthTokenExpiresIn() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return DefaultAuthTokenExpiresIn
	}

	return c.auth.TokenExpiresIn
}

// AuthTokenRefreshExpiresIn returns the duration of time the refresh token is
// valid.
func (c *Config) AuthTokenRefreshExpiresIn() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return DefaultAuthTokenRefreshExpiresIn
	}

	return c.auth.TokenRefreshExpiresIn
}

// AuthTokenIssuer returns the name of the service issuing the token.
func (c *Config) AuthTokenIssuer() string {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return DefaultAuthTokenIssuer
	}

	return c.auth.TokenIssuer
}

// AuthUpdateInterval returns the frequency at which authentication data is
// updated.
func (c *Config) AuthUpdateInterval() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return DefaultAuthUpdateInterval
	}

	return c.auth.UpdateInterval
}

// AuthIdentityDomain returns the identity domain.
func (c *Config) AuthIdentityDomain() string {
	c.RLock()
	defer c.RUnlock()

	if c.auth == nil {
		return DefaultAuthIdentityDomain
	}

	return c.auth.IdentityDomain
}

// SetAuth applies authentication configuration data to the configuration.
func (c *Config) SetAuthTokenJWKS(jwks map[string]*rsa.PublicKey) {
	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(&jwks); err != nil {
		os.Stderr.WriteString("unable to set JWKS data in configuration")
	}

	c.Lock()
	defer c.Unlock()

	c.auth.TokenJWKS = buf.String()
}
