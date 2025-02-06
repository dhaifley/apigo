package repo

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/metric"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.opentelemetry.io/otel/trace"
)

var gitClients = make(map[string]*gitClient)

var gitLock = sync.RWMutex{}

// gitClient values are used for interacting with git repositories.
type gitClient struct {
	cfg                *Config
	username, password string
	s                  storage.Storer
	fs                 billy.Filesystem
	r                  *git.Repository
	metric             metric.Recorder
	tracer             trace.Tracer
}

// newGitClient creates a new git repository client.
func newGitClient(username, password string,
	cfg *Config,
	metric metric.Recorder,
	tracer trace.Tracer,
) (*gitClient, error) {
	return &gitClient{
		username: username,
		password: password,
		cfg:      cfg,
		s:        memory.NewStorage(),
		fs:       memfs.New(),
		metric:   metric,
		tracer:   tracer,
	}, nil
}

// clone creates or updates the repository.
func (c *gitClient) clone(ctx context.Context) (*git.Repository, error) {
	if c.r == nil {
		opt := &git.CloneOptions{
			URL: c.cfg.URL,
		}

		if c.username != "" || c.password != "" {
			opt.Auth = &http.BasicAuth{
				Username: c.username,
				Password: c.password,
			}
		}

		if c.cfg.Ref != "" {
			opt.ReferenceName = plumbing.ReferenceName(c.cfg.Ref)
		}

		r, err := git.CloneContext(ctx, c.s, c.fs, opt)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrClient,
				"unable to clone repository",
				"url", c.cfg.URL)
		}

		c.r = r

		return r, nil
	}

	opt := &git.FetchOptions{}

	if c.username != "" || c.password != "" {
		opt.Auth = &http.BasicAuth{
			Username: c.username,
			Password: c.password,
		}
	}

	if err := c.r.FetchContext(ctx, opt); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil, errors.Wrap(err, errors.ErrClient,
				"unable to fetch repository",
				"url", c.cfg.URL)
		}
	}

	return c.r, nil
}

// List retrieves a directory listing from the repository.
func (c *gitClient) List(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "git",
		c.cfg, dirPath, "list")

	r, err := c.clone(ctx)
	if err != nil {
		finish(err)

		return nil, err
	}

	h, err := r.Head()
	if err != nil {
		err = errors.Wrap(err, errors.ErrClient,
			"unable to get repository commit hash",
			"url", c.cfg.URL)

		finish(err)

		return nil, err
	}

	commit := h.Hash().String()

	fis, err := c.fs.ReadDir(path.Join(c.cfg.Path, dirPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
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

	res := make([]Item, 0, len(fis))

	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}

		mt := "text/plain"

		switch filepath.Ext(fi.Name()) {
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

		t := "file"

		if fi.IsDir() {
			t = "dir"
		}

		res = append(res, Item{
			Mimetype: mt,
			Path:     path.Join(c.cfg.Path, dirPath, fi.Name()),
			Size:     int(fi.Size()),
			Type:     t,
			Commit:   commit,
		})
	}

	finish(nil)

	return res, nil
}

// ListAll retrieves a tree listing, recursively, from the repository.
func (c *gitClient) ListAll(ctx context.Context,
	dirPath string,
) ([]Item, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "git",
		c.cfg, dirPath, "listAll")

	r, err := c.clone(ctx)
	if err != nil {
		finish(err)

		return nil, err
	}

	h, err := r.Head()
	if err != nil {
		err = errors.Wrap(err, errors.ErrClient,
			"unable to get repository commit hash",
			"url", c.cfg.URL)

		finish(err)

		return nil, err
	}

	commit := h.Hash().String()

	res, err := c.listAll(ctx, dirPath, commit)
	if err != nil {
		finish(err)

		return nil, err
	}

	finish(nil)

	return res, nil
}

func (c *gitClient) listAll(ctx context.Context,
	dirPath, commit string,
) ([]Item, error) {
	fis, err := c.fs.ReadDir(path.Join(c.cfg.Path, dirPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = errors.Wrap(err, errors.ErrNotFound,
				"repository directory not found",
				"path", dirPath)
		} else {
			err = errors.Wrap(err, errors.ErrClient,
				"unable to list directory contents",
				"path", dirPath)
		}

		return nil, err
	}

	res := make([]Item, 0, len(fis))

	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), ".") || fi.Name() == "version" {
			continue
		}

		mt := "text/plain"

		switch filepath.Ext(fi.Name()) {
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

		t := "file"

		if fi.IsDir() {
			t = "dir"

			rs, err := c.listAll(ctx, path.Join(dirPath, fi.Name()), commit)
			if err != nil {
				return nil, err
			}

			res = append(res, rs...)
		}

		res = append(res, Item{
			Mimetype: mt,
			Path:     path.Join(c.cfg.Path, dirPath, fi.Name()),
			Size:     int(fi.Size()),
			Type:     t,
			Commit:   commit,
		})
	}

	return res, nil
}

// Get retrieves file contents from the repository.
func (c *gitClient) Get(ctx context.Context,
	filePath string,
) ([]byte, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "git",
		c.cfg, filePath, "get")

	if _, err := c.clone(ctx); err != nil {
		finish(err)

		return nil, err
	}

	f, err := c.fs.Open(path.Join(c.cfg.Path, filePath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
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

	defer f.Close()

	buf, err := io.ReadAll(f)
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
func (c *gitClient) Commit(ctx context.Context) (string, error) {
	_, finish := startRepoSpan(ctx, c.metric, c.tracer, "git",
		c.cfg, "main", "commit")

	r, err := c.clone(ctx)
	if err != nil {
		finish(err)

		return "", err
	}

	h, err := r.Head()
	if err != nil {
		err = errors.Wrap(err, errors.ErrClient,
			"unable to get repository commit hash",
			"url", c.cfg.URL)

		finish(err)

		return "", err
	}

	return h.Hash().String(), nil
}
