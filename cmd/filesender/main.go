package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"codeberg.org/filesender/filesender-next/config"
	"codeberg.org/filesender/filesender-next/handlers"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed public/*
var embeddedPublicFiles embed.FS

//go:embed templates/*
var embeddedTemplateFiles embed.FS

func main() {
	// Initialise database
	db, err := config.InitDB(config.GetEnv("DATABASE_PATH", "filesender.db"))
	if err != nil {
		log.Fatalf("Failed initialising database: %v", err)
		return
	}
	defer db.Close()

	// Initialise handler, pass embedded template files
	handlers.Init(embeddedTemplateFiles)

	router := http.NewServeMux()

	// API endpoints
	router.HandleFunc("GET /api/files/count", handlers.CountFilesAPIHandler(db))

	// Page handlers
	router.HandleFunc("GET /file-count", handlers.CountFilesTemplateHandler(db))

	// Serve static files
	subFS, err := fs.Sub(embeddedPublicFiles, "public")
	if err != nil {
		panic(err)
	}
	fs := http.FileServer(http.FS(subFS))
	router.Handle("GET /", http.StripPrefix("/", fs))

	addr := config.GetEnv("LISTEN", "127.0.0.1:8080")
	log.Println("HTTP server listening on " + addr)

	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Printf("Error running HTTP server: %v", err)
	}
}
