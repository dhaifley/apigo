package repo

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/metric"
	"github.com/ktrysmt/go-bitbucket"
	"go.opentelemetry.io/otel/trace"
)

// bitBucketClient values are used for interacting with BitBucket repositories.
type bitBucketClient struct {
	cfg    *Config
	cli    *bitbucket.Client
	metric metric.Recorder
	tracer trace.Tracer
}

// newBitBucketClient creates a new BitBucket repository client.
func newBitBucketClient(username, password string,
	cfg *Config,
	metric metric.Recorder,
	tracer trace.Tracer,
) (*bitBucketClient, error) {
	cli := bitbucket.NewBasicAuth(username, password)

	return &bitBucketClient{
		cfg:    cfg,
		cli:    cli,
		metric: metric,
		tracer: tracer,
	}, nil
}

// ListAll retrieves a tree listing, recursively, from the repository.
func (c *bitBucketClient) ListAll(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "bitbucket",
		c.cfg, dirPath, "list")

	opt := &bitbucket.RepositoryFilesOptions{
		Owner:    c.cfg.Owner,
		RepoSlug: c.cfg.Repo,
		Ref:      c.cfg.Ref,
		Path:     path.Join(c.cfg.Path, dirPath),
		MaxDepth: 9999,
	}

	fs, err := c.cli.Repositories.Repository.ListFiles(opt)
	if err != nil {
		if errors.ErrorHas(err, "404 Not Found") {
			err = errors.Wrap(err, errors.ErrNotFound,
				"repository directory not found",
				"path", dirPath)
		} else {
			err = errors.Wrap(err, errors.ErrClient,
				"unable to list repository directory contents",
				"path", dirPath)
		}

		finish(err)

		return nil, err
	}

	res := make([]Item, 0, len(fs))

	for _, f := range fs {
		if strings.HasPrefix(filepath.Base(f.Path), ".") {
			continue
		}

		commit := ""

		for k, v := range f.Commit {
			if vs, ok := v.(string); ok && k == "hash" {
				commit = vs

				break
			}
		}

		res = append(res, Item{
			Attributes: f.Attributes,
			Mimetype:   f.Mimetype,
			Path:       f.Path,
			Size:       f.Size,
			Type:       f.Type,
			Commit:     commit,
		})
	}

	finish(nil)

	return res, nil
}

// List retrieves a directory listing from the repository.
func (c *bitBucketClient) List(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "bitbucket",
		c.cfg, dirPath, "list")

	opt := &bitbucket.RepositoryFilesOptions{
		Owner:    c.cfg.Owner,
		RepoSlug: c.cfg.Repo,
		Ref:      c.cfg.Ref,
		Path:     path.Join(c.cfg.Path, dirPath),
	}

	fs, err := c.cli.Repositories.Repository.ListFiles(opt)
	if err != nil {
		if errors.ErrorHas(err, "404 Not Found") {
			err = errors.Wrap(err, errors.ErrNotFound,
				"repository directory not found",
				"path", dirPath)
		} else {
			err = errors.Wrap(err, errors.ErrClient,
				"unable to list repository directory contents",
				"path", dirPath)
		}

		finish(err)

		return nil, err
	}

	res := make([]Item, 0, len(fs))

	for _, f := range fs {
		if strings.HasPrefix(filepath.Base(f.Path), ".") {
			continue
		}

		commit := ""

		for k, v := range f.Commit {
			if vs, ok := v.(string); ok && k == "hash" {
				commit = vs

				break
			}
		}

		res = append(res, Item{
			Attributes: f.Attributes,
			Mimetype:   f.Mimetype,
			Path:       f.Path,
			Size:       f.Size,
			Type:       f.Type,
			Commit:     commit,
		})
	}

	finish(nil)

	return res, nil
}

// Get retrieves file contents from the repository.
func (c *bitBucketClient) Get(ctx context.Context,
	filePath string,
) ([]byte, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "bitbucket",
		c.cfg, filePath, "get")

	opt := &bitbucket.RepositoryBlobOptions{
		Owner:    c.cfg.Owner,
		RepoSlug: c.cfg.Repo,
		Ref:      c.cfg.Ref,
		Path:     path.Join(c.cfg.Path, filePath),
	}

	f, err := c.cli.Repositories.Repository.GetFileBlob(opt)
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

	if f != nil {
		if strings.HasPrefix(f.String(), `{"values":[{"path":"`) {
			err := errors.New(errors.ErrInvalidRequest,
				"unable to get repository file contents: "+
					"requested item is a directory",
				"path", filePath)

			finish(err)

			return nil, err
		}

		finish(nil)

		return f.Content, nil
	}

	finish(nil)

	return nil, nil
}

// Commit retrieves the main branch commit hash from the repository.
func (c *bitBucketClient) Commit(ctx context.Context) (string, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "github",
		c.cfg, "main", "commit")

	opt := &bitbucket.RepositoryBranchOptions{
		Owner:      c.cfg.Owner,
		RepoSlug:   c.cfg.Repo,
		BranchName: "main",
	}

	r, err := c.cli.Repositories.Repository.GetBranch(opt)
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

	commit := ""

	for k, v := range r.Target {
		if vs, ok := v.(string); ok && k == "hash" {
			commit = vs

			break
		}
	}

	return commit, nil
}
