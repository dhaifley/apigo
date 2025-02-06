// Package repo is used for accessing various online git repositories.
package repo

import (
	"context"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/dhaifley/apid/errors"
	"github.com/dhaifley/apid/metric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Client values are used to interact with a git repository.
type Client interface {
	List(ctx context.Context, dirPath string) ([]Item, error)
	ListAll(ctx context.Context, dirPath string) ([]Item, error)
	Get(ctx context.Context, filePath string) ([]byte, error)
	Commit(ctx context.Context) (string, error)
}

// Item values represent a single item in a repository.
type Item struct {
	Path       string   `json:"path"`
	Attributes []string `json:"attributes"`
	Mimetype   string   `json:"mimetype"`
	Size       int      `json:"size"`
	Type       string   `json:"type"`
	Commit     string   `json:"commit"`
}

// Config values represent configuration indicating a specific git repository.
type Config struct {
	URL   string `json:"url"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Path  string `json:"path"`
	Ref   string `json:"ref"`
}

// New is used to create a new repo client from a specified URL.
func NewClient(repoURL string,
	metric metric.Recorder,
	tracer trace.Tracer,
) (Client, error) {
	if metric == nil || (reflect.ValueOf(metric).Kind() == reflect.Ptr &&
		reflect.ValueOf(metric).IsNil()) {
		metric = nil
	}

	if tracer == nil || (reflect.ValueOf(tracer).Kind() == reflect.Ptr &&
		reflect.ValueOf(tracer).IsNil()) {
		tracer = nil
	}

	u, err := url.Parse(repoURL)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrClient,
			"invalid repository URL",
			"url", repoURL)
	}

	switch u.Scheme {
	case "bitbucket":
		if u.User == nil {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: no user information")
		}

		password, ok := u.User.Password()
		if !ok {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: no access token")
		}

		username := u.User.Username()

		cfg := &Config{Owner: u.Host}

		pe := strings.Split(strings.Trim(u.Path, "/"), "/")

		if len(pe) < 1 || pe[0] == "" {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: missing repository")
		}

		cfg.Repo = pe[0]

		if len(pe) > 1 {
			cfg.Path = strings.Join(pe[1:], "/")
		}

		cfg.Ref = u.Fragment

		return newBitBucketClient(username, password, cfg, metric, tracer)
	case "github":
		if u.User == nil {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: no user information")
		}

		password, ok := u.User.Password()
		if !ok {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: no access token")
		}

		cfg := &Config{Owner: u.Host}

		pe := strings.Split(strings.Trim(u.Path, "/"), "/")

		if len(pe) < 1 || pe[0] == "" {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: missing repository")
		}

		cfg.Repo = pe[0]

		if len(pe) > 1 {
			cfg.Path = strings.Join(pe[1:], "/")
		}

		cfg.Ref = u.Fragment

		return newGitHubClient(password, cfg, metric, tracer)
	case "test":
		if u.User == nil {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: no user information")
		}

		password, ok := u.User.Password()
		if !ok {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: no access token")
		}

		username := u.User.Username()

		cfg := &Config{Owner: u.Host}

		pe := strings.Split(strings.Trim(u.Path, "/"), "/")

		if len(pe) < 1 || pe[0] == "" {
			return nil, errors.New(errors.ErrClient,
				"invalid repository URL: missing repository")
		}

		cfg.Repo = pe[0]

		if len(pe) > 1 {
			cfg.Path = strings.Join(pe[1:], "/")
		}

		cfg.Ref = u.Fragment

		return newTestClient(username, password, cfg, metric, tracer)
	default:
		gitLock.RLock()

		if gc, ok := gitClients[u.String()]; ok {
			gitLock.RUnlock()

			return gc, nil
		}

		gitLock.RUnlock()

		username := u.User.Username()

		password, _ := u.User.Password()

		u.User = nil

		cfg := &Config{URL: u.String()}

		gc, err := newGitClient(username, password, cfg, metric, tracer)
		if err != nil {
			return nil, err
		}

		gitLock.Lock()

		gitClients[u.String()] = gc

		gitLock.Unlock()

		return gc, nil
	}
}

// testClient values are used for testing repository interactions.
type testClient struct {
	u, p   string
	cfg    *Config
	metric metric.Recorder
	tracer trace.Tracer
}

// newTestClient creates a new test repository client.
func newTestClient(username, password string,
	cfg *Config,
	metric metric.Recorder,
	tracer trace.Tracer,
) (*testClient, error) {
	return &testClient{
		u:      username,
		p:      password,
		cfg:    cfg,
		metric: metric,
		tracer: tracer,
	}, nil
}

// List retrieves a directory listing from the repository.
func (c *testClient) List(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "test",
		c.cfg, dirPath, "list")

	defer finish(nil)

	return []Item{{Path: dirPath}}, nil
}

// List retrieves a directory listing from the repository.
func (c *testClient) ListAll(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "test",
		c.cfg, dirPath, "listAll")

	defer finish(nil)

	return []Item{{Path: filepath.Join(dirPath, "test")}}, nil
}

// Get retrieves file contents from the repository.
func (c *testClient) Get(ctx context.Context,
	filePath string,
) ([]byte, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "test",
		c.cfg, filePath, "get")

	defer finish(nil)

	return []byte(filePath), nil
}

// Commit retrieves the main branch commit hash from the repository.
func (c *testClient) Commit(ctx context.Context) (string, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "test",
		c.cfg, "main", "commit")

	defer finish(nil)

	return "", nil
}

// startRepoSpan starts a cache tracing span. It returns an updated context,
// and a closing function.
func startRepoSpan(ctx context.Context,
	metric metric.Recorder,
	tracer trace.Tracer,
	system string,
	cfg *Config,
	repoPath, name string,
) (context.Context, func(err error)) {
	start := time.Now()

	if tracer == nil {
		return ctx, func(err error) {
			if metric != nil {
				metric.RecordDuration(ctx, "repo_latency",
					time.Since(start), "operation:"+name, "system:"+system)
			}
		}
	}

	attrs := []attribute.KeyValue{
		attribute.String("repo.system", system),
		attribute.String("repo.path", repoPath),
	}

	if cfg != nil {
		attrs = append(attrs,
			attribute.String("repo.config.owner", cfg.Owner),
			attribute.String("repo.config.repo", cfg.Repo),
			attribute.String("repo.config.path", cfg.Path),
			attribute.String("repo.config.ref", cfg.Ref),
		)
	}

	ctx, span := tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)

	return ctx, func(err error) {
		if err != nil && reflect.ValueOf(err).Kind() == reflect.Ptr &&
			reflect.ValueOf(err).IsNil() {
			err = nil
		}

		if span != nil {
			if err != nil {
				span.SetStatus(codes.Error, name+" failed")
				span.RecordError(err)
			}

			span.End()
		}

		if metric != nil {
			metric.RecordDuration(ctx, "repo_latency",
				time.Since(start), "operation:"+name, "system:"+system)
		}
	}
}
