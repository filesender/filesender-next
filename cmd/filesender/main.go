package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"time"

	"codeberg.org/filesender/filesender-next/internal/assets"
	"codeberg.org/filesender/filesender-next/internal/config"
	"codeberg.org/filesender/filesender-next/internal/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialise database
	db, err := config.InitDB(path.Join(config.GetEnv("STATE_DIRECTORY", "."), "filesender.db"))
	if err != nil {
		slog.Error("Failed initialising database", "error", err)
		return
	}
	defer db.Close()

	// Initialise handler, pass embedded template files
	handlers.Init(assets.EmbeddedTemplateFiles)

	router := http.NewServeMux()

	// API endpoints
	router.HandleFunc("GET /api/files/count", handlers.CountFilesAPIHandler(db))

	// Page handlers
	router.HandleFunc("GET /file-count", handlers.CountFilesTemplateHandler(db))

	// Serve static files
	subFS, err := fs.Sub(assets.EmbeddedPublicFiles, "public")
	if err != nil {
		panic(err)
	}
	fs := http.FileServer(http.FS(subFS))
	router.Handle("GET /", http.StripPrefix("/", fs))

	// Setup server
	addr := config.GetEnv("LISTEN", "127.0.0.1:8080")
	s := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	slog.Info("HTTP server listening on " + addr)
	err = s.ListenAndServe()
	if err != nil {
		slog.Error("Error running HTTP server", "error", err)
	}
}
