package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:out
var distFS embed.FS

// Assets returns the fs.FS for the exported Next.js frontend assets.
func Assets() (fs.FS, error) {
	return fs.Sub(distFS, "out")
}
