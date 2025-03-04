package main

import (
	"log"
	"net/http"
	"os"

	"codeberg.org/filesender/filesender-next/src/config"
	"codeberg.org/filesender/filesender-next/src/handlers"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load .env file if any exists
	config.LoadEnv()

	db_path, exists := os.LookupEnv("DATABASE_PATH")
	if !exists {
		log.Fatalf("Database path not set in env")
		return
	}

	// Initialise database
	db, err := config.InitDB(db_path)
	if err != nil {
		log.Fatalf("Failed initialising database: %v", err)
		return
	}
	defer db.Close()

	router := mux.NewRouter()

	// API endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/files/count", handlers.CountFilesApiHandler(db)).Methods("GET")

	// Page handlers
	router.HandleFunc("/file-count", handlers.CountFilesTemplateHandler(db)).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("./public"))
	router.PathPrefix("/").Handler(fs)

	port := "8080"
	log.Fatal(http.ListenAndServe(":"+port, router))
}
