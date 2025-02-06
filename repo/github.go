package repo

import (
	"context"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/dhaifley/apid/errors"
	"github.com/dhaifley/apid/metric"
	"github.com/google/go-github/v39/github"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
)

// gitHubClient values are used for interacting with GitHub repositories.
type gitHubClient struct {
	cfg    *Config
	cli    *github.Client
	metric metric.Recorder
	tracer trace.Tracer
}

// newGitHubClient creates a new GitHub repository client.
func newGitHubClient(password string,
	cfg *Config,
	metric metric.Recorder,
	tracer trace.Tracer,
) (*gitHubClient, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: password})

	c := oauth2.NewClient(context.Background(), ts)

	cli := github.NewClient(c)

	return &gitHubClient{
		cfg:    cfg,
		cli:    cli,
		metric: metric,
		tracer: tracer,
	}, nil
}

// List retrieves a directory listing from the repository.
func (c *gitHubClient) List(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "github",
		c.cfg, dirPath, "list")

	opt := &github.RepositoryContentGetOptions{
		Ref: c.cfg.Ref,
	}

	fc, dc, _, err := c.cli.Repositories.GetContents(ctx, c.cfg.Owner,
		c.cfg.Repo, path.Join(c.cfg.Path, dirPath), opt)
	if err != nil {
		if errors.ErrorHas(err, "404 Not Found") {
			err = errors.Wrap(err, errors.ErrNotFound,
				"repository directory not found",
				"path", dirPath)
		} else {
			err = errors.Wrap(err, errors.ErrClient,
				"unable to list directory contents",
				"path", dirPath)
		}

		finish(err)

		return nil, err
	}

	if fc != nil {
		err = errors.Wrap(err, errors.ErrNotFound,
			"repository directory not found",
			"path", dirPath)

		finish(err)

		return nil, err
	}

	res := make([]Item, 0, len(dc))

	for _, rc := range dc {
		if strings.HasPrefix(filepath.Base(rc.GetPath()), ".") {
			continue
		}

		mt := "text/plain"

		switch filepath.Ext(rc.GetPath()) {
		case ".zip":
			mt = "application/zip"
		case ".yaml", ".yml":
			mt = "application/yaml"
		case ".json":
			mt = "application/json"
		case ".toml":
			mt = "application/toml"
		case ".xml":
			mt = "application/xml"
		case ".sh":
			mt = "application/x-sh"
		case ".exe":
			mt = "application/ms-dos"
		}

		res = append(res, Item{
			Mimetype: mt,
			Path:     rc.GetPath(),
			Size:     rc.GetSize(),
			Type:     rc.GetType(),
			Commit:   rc.GetSHA(),
		})
	}

	finish(nil)

	return res, nil
}

// ListAll retrieves a tree listing, recursively, from the repository.
func (c *gitHubClient) ListAll(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "github",
		c.cfg, "/", "listAll")

	t, _, err := c.cli.Git.GetTree(ctx, c.cfg.Owner,
		c.cfg.Repo, "main", true)
	if err != nil {
		err = errors.Wrap(err, errors.ErrClient,
			"unable to get repository tree")

		finish(err)

		return nil, err
	}

	res := make([]Item, 0, len(t.Entries))

TreeLoop:
	for _, te := range t.Entries {
		if !strings.HasPrefix(te.GetPath(), dirPath) ||
			strings.HasSuffix(te.GetPath(), "/version") {
			continue TreeLoop
		}

		seg := strings.Split(te.GetPath(), "/")

		for _, sp := range seg {
			if strings.HasPrefix(sp, ".") {
				continue TreeLoop
			}
		}

		mt := "text/plain"

		switch {
		case strings.HasSuffix(te.GetPath(), ".zip"):
			mt = "application/zip"
		case strings.HasSuffix(te.GetPath(), ".yaml"),
			strings.HasSuffix(te.GetPath(), ".yml"):
			mt = "application/yaml"
		case strings.HasSuffix(te.GetPath(), ".json"):
			mt = "application/json"
		case strings.HasSuffix(te.GetPath(), ".toml"):
			mt = "application/toml"
		case strings.HasSuffix(te.GetPath(), ".xml"):
			mt = "application/xml"
		case strings.HasSuffix(te.GetPath(), ".sh"):
			mt = "application/x-sh"
		case strings.HasSuffix(te.GetPath(), ".exe"):
			mt = "application/ms-dos"
		}

		ft := "file"

		if te.GetType() == "tree" {
			ft = "dir"
		}

		res = append(res, Item{
			Mimetype: mt,
			Path:     te.GetPath(),
			Size:     te.GetSize(),
			Type:     ft,
			Commit:   te.GetSHA(),
		})
	}

	finish(nil)

	return res, nil
}

// Get retrieves file contents from the repository.
func (c *gitHubClient) Get(ctx context.Context,
	filePath string,
) ([]byte, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "github",
		c.cfg, filePath, "get")

	opt := &github.RepositoryContentGetOptions{
		Ref: c.cfg.Ref,
	}

	r, _, err := c.cli.Repositories.DownloadContents(ctx,
		c.cfg.Owner, c.cfg.Repo, filePath, opt)
	if err != nil {
		if errors.ErrorHas(err, "404 Not Found") {
			err = errors.Wrap(err, errors.ErrNotFound,
				"repository file not found",
				"path", filePath)
		} else {
			err = errors.Wrap(err, errors.ErrClient,
				"unable to get repository file contents",
				"path", filePath)
		}

		finish(err)

		return nil, err
	}

	defer r.Close()

	buf, err := io.ReadAll(r)
	if err != nil {
		err = errors.Wrap(err, errors.ErrClient,
			"unable to read repository file contents",
			"path", filePath)

		finish(err)

		return nil, err
	}

	if buf != nil {
		finish(nil)

		return buf, nil
	}

	finish(nil)

	return nil, nil
}

// Commit retrieves the main branch commit hash from the repository.
func (c *gitHubClient) Commit(ctx context.Context) (string, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "github",
		c.cfg, "main", "commit")

	r, _, err := c.cli.Repositories.GetBranch(ctx,
		c.cfg.Owner, c.cfg.Repo, "main", true)
	if err != nil {
		if errors.ErrorHas(err, "404 Not Found") {
			err = errors.Wrap(err, errors.ErrNotFound,
				"repository main branch not found")
		} else {
			err = errors.Wrap(err, errors.ErrClient,
				"unable to get repository main branch")
		}

		finish(err)

		return "", err
	}

	finish(nil)

	if r.Commit == nil || r.Commit.SHA == nil {
		return "", nil
	}

	return *r.Commit.SHA, nil
}
