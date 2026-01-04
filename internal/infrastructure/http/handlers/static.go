package handlers

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// SPAHandler serves static files with SPA fallback.
// If a file is not found, it serves index.html for client-side routing.
type SPAHandler struct {
	staticFS   fs.FS
	staticPath string
}

// NewSPAHandler creates a new SPA handler.
// staticFS is the embedded filesystem.
// staticPath is the subdirectory containing the static files (e.g., "dist").
func NewSPAHandler(staticFS fs.FS, staticPath string) *SPAHandler {
	return &SPAHandler{
		staticFS:   staticFS,
		staticPath: staticPath,
	}
}

// ServeHTTP implements http.Handler.
func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}
	upath = path.Clean(upath)

	// Build the full path within the embedded filesystem
	fullPath := path.Join(h.staticPath, upath)

	// Try to open the file
	file, err := h.staticFS.Open(fullPath)
	if err != nil {
		// File not found - serve index.html for SPA routing
		h.serveIndex(w, r)
		return
	}
	defer file.Close()

	// Check if it's a directory
	stat, err := file.Stat()
	if err != nil {
		h.serveIndex(w, r)
		return
	}

	// If it's a directory, try to serve index.html from that directory
	if stat.IsDir() {
		indexPath := path.Join(fullPath, "index.html")
		indexFile, err := h.staticFS.Open(indexPath)
		if err != nil {
			h.serveIndex(w, r)
			return
		}
		defer indexFile.Close()
		h.serveFile(w, r, indexPath)
		return
	}

	// Serve the file
	h.serveFile(w, r, fullPath)
}

// serveIndex serves the root index.html for SPA fallback.
func (h *SPAHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := path.Join(h.staticPath, "index.html")
	h.serveFile(w, r, indexPath)
}

// serveFile serves a file from the embedded filesystem.
func (h *SPAHandler) serveFile(w http.ResponseWriter, _ *http.Request, filePath string) {
	file, err := h.staticFS.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info for Content-Length
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set content type based on extension
	contentType := getContentType(filePath)
	w.Header().Set("Content-Type", contentType)

	// Set cache headers for assets
	if strings.HasPrefix(filePath, path.Join(h.staticPath, "assets")) {
		// Assets have hashed filenames, cache for 1 year
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if strings.HasSuffix(filePath, "index.html") {
		// index.html should not be cached
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// Read and write content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", string(rune(stat.Size())))
	_, _ = w.Write(content)
}

// getContentType returns the MIME type for a file based on its extension.
func getContentType(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}
