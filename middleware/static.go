package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/jasonsoft/napnap"
)

// Static is a middleware handler that serves static files in the given directory/filesystem.
type Static struct {
	// Dir is the directory to serve static files from
	Dir http.FileSystem
	// Prefix is the optional prefix used to serve the static directory content
	Prefix string
	// IndexFile defines which file to serve as index if it exists.
	IndexFile string
}

// NewStatic returns a new instance of Static
func NewStatic(dir string) *Static {
	directory := http.Dir(dir)
	return &Static{
		Dir:       directory,
		Prefix:    "",
		IndexFile: "index.html",
	}
}

// Invoke function is a middleware entry
func (s *Static) Invoke(c *napnap.Context, next napnap.HandlerFunc) {
	r := c.Request
	if r.Method != "GET" && r.Method != "HEAD" {
		_ = next(c)
		return
	}
	file := r.URL.Path
	// if we have a prefix, filter requests by stripping the prefix
	if s.Prefix != "" {
		if !strings.HasPrefix(file, s.Prefix) {
			_ = next(c)
			return
		}
		file = file[len(s.Prefix):]
		if file != "" && file[0] != '/' {
			_ = next(c)
			return
		}
	}
	f, err := s.Dir.Open(file)
	if err != nil {
		// discard the error?
		_ = next(c)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		_ = next(c)
		return
	}

	// try to serve index file
	if fi.IsDir() {
		// redirect if missing trailing slash
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(c.Writer, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		file = path.Join(file, s.IndexFile)
		f, err = s.Dir.Open(file)
		if err != nil {
			_ = next(c)
			return
		}
		defer f.Close()

		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			_ = next(c)
			return
		}
	}

	http.ServeContent(c.Writer, r, file, fi.ModTime(), f)
}
