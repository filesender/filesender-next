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

func main() {
	addr := flag.String("listen", "127.0.0.1:8080", "specify the LISTEN address")
	flag.Parse()

	var authModule auth.Auth
	authModule = &auth.ProxyAuth{}
	if os.Getenv("FILESENDER_AUTH_METHOD") == "dummy" {
		slog.Info("Using `dummy` authentication method")
		authModule = &auth.DummyAuth{}
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

	// Initialise handler, pass embedded template files
	handlers.Init(assets.EmbeddedTemplateFiles)

	router := http.NewServeMux()
	// API endpoints
	router.HandleFunc("POST /api/upload", handlers.UploadAPI(authModule, stateDir, maxUploadSize))
	router.HandleFunc("PATCH /api/upload/{fileID}", handlers.ChunkedUploadAPI(authModule, stateDir, maxUploadSize))

	stateDirFS := http.FileServer(http.Dir(stateDir))
	router.Handle("/download/", http.StripPrefix("/download/", stateDirFS))

	// Page handlers
	router.HandleFunc("GET /{$}", handlers.UploadTemplate(authModule))
	router.HandleFunc("GET /view/{userID}/{fileID}", handlers.GetDownloadTemplate(stateDir))

	// Serve static files
	subFS, err := fs.Sub(assets.EmbeddedPublicFiles, "public")
	if err != nil {
		slog.Error("Failed serving static files", "error", err)
		os.Exit(1)
	}
	fs := http.FileServer(http.FS(subFS))
	router.Handle("/", http.StripPrefix("/", fs))

	// Setup server
	s := &http.Server{
		Addr:           *addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	slog.Info("HTTP server listening on " + *addr)
	err = s.ListenAndServe()
	if err != nil {
		slog.Error("Error running HTTP server", "error", err)
		os.Exit(1)
	}
}
