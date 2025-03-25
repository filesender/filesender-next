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
	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/logging"
)

func main() {
	enableDebug := flag.Bool("d", false, "enable DEBUG output")
	addr := flag.String("listen", "127.0.0.1:8080", "specify the LISTEN address")
	flag.Parse()

	// Set log level if debug
	if *enableDebug {
		logging.SetLogLevel(slog.LevelDebug)
	}

	// Get max upload size (per chunk)
	maxUploadSizeStr := os.Getenv("MAX_UPLOAD_SIZE")
	if maxUploadSizeStr == "" {
		slog.Info("environment variable \"MAX_UPLOAD_SIZE\" is not set, using default: 2147483648 (2GB)")
		maxUploadSizeStr = "2147483648"
	}
	maxUploadSize, err := strconv.ParseInt(maxUploadSizeStr, 10, 0)
	if err != nil {
		slog.Error("Failed converting \"MAX_UPLOAD_SIZE\" to int", "error", err)
		os.Exit(1)
	}

	// Initialise database
	stateDir := os.Getenv("STATE_DIRECTORY")
	if stateDir == "" {
		slog.Error("environment variable \"STATE_DIRECTORY\" not set")
		os.Exit(1)
	}
	err = os.MkdirAll(stateDir, os.ModePerm)
	if err != nil {
		slog.Error("Failed creating state directory", "error", err)
	}

	// Initialise handler, pass embedded template files
	handlers.Init(assets.EmbeddedTemplateFiles)

	router := http.NewServeMux()
	// API endpoints
	router.HandleFunc("POST /api/v1/transfers", handlers.CreateTransferAPIHandler())
	router.HandleFunc("POST /api/v1/upload", handlers.UploadAPIHandler(maxUploadSize))

	// Page handlers
	router.HandleFunc("GET /{$}", handlers.UploadTemplateHandler())
	router.HandleFunc("GET /upload/{id}", handlers.UploadDoneTemplateHandler())
	router.HandleFunc("GET /transfer/{userID}/{transferID}", handlers.GetTransferTemplateHandler())

	// Serve static files
	subFS, err := fs.Sub(assets.EmbeddedPublicFiles, "public")
	if err != nil {
		panic(err)
	}
	fs := http.FileServer(http.FS(subFS))
	router.Handle("GET /", http.StripPrefix("/", fs))

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
	}
}
