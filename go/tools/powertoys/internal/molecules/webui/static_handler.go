package webui

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticHandler handles serving static frontend files.
type StaticHandler struct {
	fileSystem fs.FS
	devMode    bool
	devPath    string
}

// NewStaticHandler creates a new static file handler.
// If WEBUI_DEV_MODE environment variable is set, serves files directly from webui/public directory.
// Otherwise, uses embedded filesystem.
func NewStaticHandler() *StaticHandler {
	devMode := os.Getenv("WEBUI_DEV_MODE") == "1" || os.Getenv("WEBUI_DEV_MODE") == "true"

	var handler *StaticHandler
	if devMode {
		// Development mode: serve from webui/public directory
		// Try to find webui/public relative to current working directory or executable
		devPath := findWebUIPublicDir()
		if devPath != "" {
			fmt.Printf("[StaticHandler] Development mode enabled - serving from: %s\n", devPath)
			handler = &StaticHandler{
				fileSystem: os.DirFS(devPath),
				devMode:    true,
				devPath:    devPath,
			}
		} else {
			fmt.Printf("[StaticHandler] Development mode enabled but webui/public not found, falling back to embedded assets\n")
			handler = &StaticHandler{
				fileSystem: embeddedAssets,
				devMode:    false,
			}
		}
	} else {
		// Production mode: use embedded filesystem
		handler = &StaticHandler{
			fileSystem: embeddedAssets,
			devMode:    false,
		}
	}

	return handler
}

// findWebUIPublicDir attempts to find the webui/public directory.
// It checks relative to the current working directory and common project structures.
func findWebUIPublicDir() string {
	// Try current directory
	cwd, err := os.Getwd()
	if err == nil {
		paths := []string{
			filepath.Join(cwd, "webui", "public"),
			filepath.Join(cwd, "..", "webui", "public"),
			filepath.Join(cwd, "..", "..", "webui", "public"),
		}
		for _, path := range paths {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				absPath, _ := filepath.Abs(path)
				return absPath
			}
		}
	}

	// Try relative to executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		paths := []string{
			filepath.Join(exeDir, "webui", "public"),
			filepath.Join(exeDir, "..", "webui", "public"),
		}
		for _, path := range paths {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				absPath, _ := filepath.Abs(path)
				return absPath
			}
		}
	}

	return ""
}

// ServeFiles registers static file routes with the given mux.
func (sh *StaticHandler) ServeFiles(mux *http.ServeMux) {
	// Handle favicon.ico explicitly (browsers request it automatically)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve favicon.ico from static files
		// If not found, return 204 No Content (standard for missing favicon)
		if _, err := fs.Stat(sh.fileSystem, "favicon.ico"); err == nil {
			sh.serveFile(w, r, "favicon.ico")
			return
		}
		// Also try with public/ prefix in production mode
		if !sh.devMode {
			if _, err := fs.Stat(sh.fileSystem, "public/favicon.ico"); err == nil {
				sh.serveFile(w, r, "public/favicon.ico")
				return
			}
		}
		// Return 204 No Content instead of 404 to suppress browser errors
		w.WriteHeader(http.StatusNoContent)
	})

	// Handle root and static files
	// Note: This should be registered AFTER API routes so API routes take precedence
	// In Go's http.ServeMux, more specific routes (like /api/...) take precedence over /
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Don't handle API routes - let API routes handler take care of them
		// More specific routes registered first will match before this catch-all
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// This shouldn't happen if API routes are registered correctly,
			// but just in case, return 404
			http.NotFound(w, r)
			return
		}

		if r.URL.Path == "/" || r.URL.Path == "" {
			// Serve main.html for root
			sh.serveFile(w, r, "pages/html/main.html")
			return
		}

		// Check if it's a page request (ends with .html)
		if strings.HasSuffix(r.URL.Path, ".html") {
			// Map common routes to HTML files
			htmlPath := sh.mapRouteToHTML(r.URL.Path)
			sh.serveFile(w, r, htmlPath)
			return
		}

		// For all other paths, remove leading slash and try to find file
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			sh.serveFile(w, r, "pages/html/main.html")
			return
		}

		// Try direct path first (since embed root is "public", files are at root level)
		// But wait - embed is "public", so root contains "public" directory
		// So we need to check if path exists directly, or with public/ prefix

		// Normalize path separators (embedded FS always uses forward slashes)
		path = strings.ReplaceAll(path, "\\", "/")

		// In dev mode, files are served directly from webui/public, so path is correct as-is
		// In production mode, embed root is "public", so we need "public/" prefix
		if sh.devMode {
			// Dev mode: try path as-is first
			if _, err := fs.Stat(sh.fileSystem, path); err == nil {
				sh.serveFile(w, r, path)
				return
			}
		} else {
			// Production mode: try with public/ prefix first
			publicPath := "public/" + path
			if _, err := fs.Stat(sh.fileSystem, publicPath); err == nil {
				sh.serveFile(w, r, publicPath)
				return
			}
			// Fallback: try without public/ prefix
			if _, err := fs.Stat(sh.fileSystem, path); err == nil {
				sh.serveFile(w, r, path)
				return
			}
		}

		// Try with common prefixes
		if sh.devMode {
			// Dev mode: no public/ prefix needed
			for _, prefix := range []string{"pages/html/", "pages/js/", "pages/css/", "molecules/js/", "molecules/css/", "atoms/css/", "atoms/js/", "css/", "templates/css/", "templates/html/"} {
				fullPath := prefix + path
				if _, err := fs.Stat(sh.fileSystem, fullPath); err == nil {
					sh.serveFile(w, r, fullPath)
					return
				}
			}
		} else {
			// Production mode: try with public/ prefix
			for _, prefix := range []string{"public/pages/html/", "public/pages/js/", "public/pages/css/", "public/molecules/js/", "public/molecules/css/", "public/atoms/css/", "public/atoms/js/", "public/css/", "public/templates/css/", "public/templates/html/"} {
				fullPath := prefix + path
				if _, err := fs.Stat(sh.fileSystem, fullPath); err == nil {
					sh.serveFile(w, r, fullPath)
					return
				}
			}
			// Fallback: try without public/ prefix
			for _, prefix := range []string{"pages/html/", "pages/js/", "pages/css/", "molecules/js/", "molecules/css/", "atoms/css/", "atoms/js/", "css/", "templates/css/", "templates/html/"} {
				fullPath := prefix + path
				if _, err := fs.Stat(sh.fileSystem, fullPath); err == nil {
					sh.serveFile(w, r, fullPath)
					return
				}
			}
		}

		// Not found
		fmt.Printf("[StaticHandler] File not found for path: %s\n", r.URL.Path)
		http.NotFound(w, r)
	})
}

// mapRouteToHTML maps URL routes to HTML file paths.
func (sh *StaticHandler) mapRouteToHTML(path string) string {
	// Remove leading slash and .html extension
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, ".html")

	// Map common routes (embed root is "public", so files are at "public/pages/html/...")
	switch path {
	case "", "index", "main":
		return "pages/html/main.html"
	case "pages":
		return "pages/html/pages.html"
	case "macros":
		return "pages/html/macros.html"
	case "remote-upload", "upload":
		return "pages/html/remote-upload.html"
	case "rcu":
		return "pages/html/rcu.html"
	default:
		// Try pages/html/{path}.html
		return "pages/html/" + path + ".html"
	}
}

// serveFile serves a specific file from the filesystem (embedded or dev mode).
func (sh *StaticHandler) serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// In production mode, embed root is "public", so add "public/" prefix if not present
	// In dev mode, files are already in the correct path relative to webui/public
	if !sh.devMode && !strings.HasPrefix(filePath, "public/") {
		filePath = "public/" + filePath
	}

	// IMPORTANT: Embedded filesystems always use forward slashes, even on Windows
	// Convert any backslashes to forward slashes (filepath.Clean converts to OS-specific separators)
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	filePath = strings.TrimPrefix(filePath, "./") // Remove any leading ./

	// Debug: Log what we're trying to access
	fmt.Printf("[StaticHandler] Attempting to serve file: %s (request: %s)\n", filePath, r.URL.Path)

	// Check if file exists first
	if _, err := fs.Stat(sh.fileSystem, filePath); err != nil {
		fmt.Printf("[StaticHandler] File not found: %s, error: %v\n", filePath, err)
		// Try to list what's in the filesystem root for debugging
		if entries, listErr := fs.ReadDir(sh.fileSystem, "."); listErr == nil {
			fmt.Printf("[StaticHandler] Filesystem root contents:\n")
			for _, entry := range entries {
				fmt.Printf("  - %s (dir: %v)\n", entry.Name(), entry.IsDir())
			}
		}
		http.NotFound(w, r)
		return
	}

	// Read file from embedded filesystem
	data, err := fs.ReadFile(sh.fileSystem, filePath)
	if err != nil {
		fmt.Printf("[StaticHandler] Error reading file %s: %v\n", filePath, err)
		http.NotFound(w, r)
		return
	}

	fmt.Printf("[StaticHandler] Successfully serving file: %s (size: %d bytes)\n", filePath, len(data))

	// Set content type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Write file content
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

