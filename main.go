package main

import (
	"log"
	"net/http"
	"os"

	"github.com/danmcfan/eco-stream/internal/handlers"
	"github.com/danmcfan/eco-stream/internal/middleware"
	"github.com/danmcfan/eco-stream/internal/minio"
	"github.com/danmcfan/eco-stream/internal/postgres"
)

func main() {
	db := postgres.CreatePostgresClient()
	defer db.Close()

	minioClient := minio.CreateMinioClient()
	minio.CreateBucket(minioClient, "default", "us-east-1")

	http.HandleFunc("/health/", handlers.HealthCheckHandler)
	http.HandleFunc("/users/", middleware.CorsMiddleware(handlers.UserHandlers(db)))
	http.HandleFunc("/files/", handlers.FileHandlers(minioClient))

	listenerURL := "localhost:8080"
	if val, ok := os.LookupEnv("LISTENER_URL"); ok {
		listenerURL = val
	}

	log.Printf("Server is listening on %s", listenerURL)
	if err := http.ListenAndServe(listenerURL, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
