package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"codeberg.org/filesender/filesender-next/config"
	"codeberg.org/filesender/filesender-next/handlers"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed public/*
var embeddedPublicFiles embed.FS

//go:embed templates/*
var embeddedTemplateFiles embed.FS

func main() {
	// Load .env file if any exists
	cnfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed loading configuration: %v", err)
		return
	}

	// Initialise database
	db, err := config.InitDB(cnfg["database"]["path"])
	if err != nil {
		log.Fatalf("Failed initialising database: %v", err)
		return
	}
	defer db.Close()

	// Initialise handler, pass embedded template files
	handlers.Init(embeddedTemplateFiles)

	router := mux.NewRouter()

	// API endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/files/count", handlers.CountFilesAPIHandler(db)).Methods("GET")

	// Page handlers
	router.HandleFunc("/file-count", handlers.CountFilesTemplateHandler(db)).Methods("GET")

	// Serve static files
	subFS, err := fs.Sub(embeddedPublicFiles, "public")
	if err != nil {
		panic(err)
	}
	fs := http.FileServer(http.FS(subFS))
	router.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	addr := cnfg["http"]["host"] + ":" + cnfg["http"]["port"]
	log.Printf("HTTP server listening on " + addr)
	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Printf("Error runngin HTTP server: %v", err)
	}
}
