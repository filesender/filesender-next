package main

import (
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"

	"codeberg.org/filesender/filesender-next/internal/assets"
	"codeberg.org/filesender/filesender-next/internal/config"
	"codeberg.org/filesender/filesender-next/internal/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	enableDebug := flag.Bool("d", false, "enable DEBUG output")
	addr := flag.String("listen", "127.0.0.1:8080", "specify the LISTEN address")

	flag.Parse()

	setLogLevel(*enableDebug)

	// Initialise database
	stateDir := os.Getenv("STATE_DIRECTORY")
	if stateDir == "" {
		slog.Error("environment variable \"STATE_DIRECTORY\" not set")
		os.Exit(1)
	}
	db, err := config.InitDB(path.Join(stateDir, "filesender.db"))
	if err != nil {
		slog.Error("Failed initialising database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialise handler, pass embedded template files
	handlers.Init(assets.EmbeddedTemplateFiles)

	router := http.NewServeMux()

	// API endpoints
	router.HandleFunc("POST /api/v1/transfers", handlers.CreateTransferAPIHandler(db))

	// Page handlers
	router.HandleFunc("GET /{$}", handlers.UploadTemplateHandler())

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

func setLogLevel(enableDebug bool) {
	logLevel := slog.LevelInfo
	if enableDebug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
	slog.Debug("Debug logging enabled")
}
