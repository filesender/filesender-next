// Package main, the starting point of the Filesender application
package main

import (
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"codeberg.org/filesender/filesender-next/internal/assets"
	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/hash"
)

func maxUploadSize() int64 {
	// parse MAX_UPLOAD_SIZE environment variable as an unsigned integer. If
	// not specified, or parsing fails, return the default
	muInt, err := strconv.ParseUint(os.Getenv("MAX_UPLOAD_SIZE"), 10, 0)
	if err != nil {
		// default = 2 GiB
		return 2 * 1024 * 1024 * 1024
	}

	return int64(muInt)
}

func wrapHandlerWithTimeout(f func(http.ResponseWriter, *http.Request)) http.Handler {
	hf := http.HandlerFunc(f)
	return http.TimeoutHandler(hf, time.Second*10, "")
}

func main() {
	addr := flag.String("listen", "127.0.0.1:8080", "specify the LISTEN address")
	flag.Parse()

	var authModule auth.Auth
	authModule = &auth.ProxyAuth{}
	if os.Getenv("FILESENDER_AUTH_METHOD") == "dummy" {
		slog.Info("Using `dummy` authentication method")
		authModule = &auth.DummyAuth{}
	}

	appRoot := os.Getenv("FILESENDER_APP_ROOT")
	if appRoot == "" {
		appRoot = "/"
	}

	maxUploadSize := maxUploadSize()
	slog.Info("MAX_UPLOAD_SIZE", "bytes", maxUploadSize)

	// Initialise database
	stateDir := os.Getenv("STATE_DIRECTORY")
	if stateDir == "" {
		slog.Error("environment variable \"STATE_DIRECTORY\" not set")
		os.Exit(1)
	}
	err := os.MkdirAll(stateDir, 0o700)
	if err != nil {
		slog.Error("Failed creating state directory", "error", err)
		os.Exit(1)
	}
	slog.Info("State directory set", "dir", stateDir)

	// Initialise hashing function
	err = hash.Init(stateDir)
	if err != nil {
		slog.Error("Failed initialising hashing", "error", err)
		os.Exit(1)
	}

	// Initialise handler, pass embedded template files
	handlers.Init(assets.EmbeddedTemplateFiles)

	router := http.NewServeMux()
	// API endpoints
	router.Handle("POST /upload", wrapHandlerWithTimeout(handlers.UploadAPI(appRoot, authModule, stateDir, maxUploadSize)))
	router.Handle("PATCH /upload/{fileID}", wrapHandlerWithTimeout(handlers.ChunkedUploadAPI(appRoot, authModule, stateDir, maxUploadSize)))

	stateDirFS := http.FileServer(http.Dir(stateDir))
	router.Handle("/download/{a}/{b}", http.StripPrefix("/download/", stateDirFS))

	// Page handlers
	router.Handle("GET /{$}", wrapHandlerWithTimeout(handlers.UploadTemplate(appRoot, authModule)))
	router.Handle("GET /view/{userID}/{fileID}", wrapHandlerWithTimeout(handlers.GetDownloadTemplate(appRoot, stateDir)))

	// Serve static files
	subFS, err := fs.Sub(assets.EmbeddedPublicFiles, "public")
	if err != nil {
		slog.Error("Failed serving static files", "error", err)
		os.Exit(1)
	}
	fs := http.FileServer(http.FS(subFS))
	withHeaders := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "js/sw.js" {
			w.Header().Set("Service-Worker-Allowed", "/")
		}
		fs.ServeHTTP(w, r)
	})
	router.Handle("/", http.StripPrefix("/", wrapHandlerWithTimeout(withHeaders)))

	// Setup server
	s := &http.Server{
		Addr:           *addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}

	slog.Info("HTTP server listening on " + *addr)
	err = s.ListenAndServe()
	if err != nil {
		slog.Error("Error running HTTP server", "error", err)
		os.Exit(1)
	}
}
