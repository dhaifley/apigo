package static

import "embed"

// FS is a file system containing static files.
//
//go:embed *
var FS embed.FS
