package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/minio/minio-go"

	internalMinio "github.com/danmcfan/eco-stream/internal/minio"
	"github.com/danmcfan/eco-stream/internal/models"
	internalRedis "github.com/danmcfan/eco-stream/internal/redis"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "All systems operational!")
}

func UserHandlers(rdb *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			id := strings.TrimPrefix(r.URL.Path, "/users/")
			if id == "" {
				listUsersHandler(rdb)(w, r)
			} else {
				retrieveUserHandler(rdb)(w, r)
			}
		case http.MethodPost:
			createUserHandler(rdb)(w, r)
		case http.MethodDelete:
			deleteUserHandler(rdb)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listUsersHandler(rdb *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := internalRedis.ListUsers(rdb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(users)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}
}

func retrieveUserHandler(rdb *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		user, err := internalRedis.RetrieveUser(rdb, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		jsonResponse, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}
}

func createUserHandler(rdb *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.CreateUser
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newUser := models.User{
			ID:       uuid.New().String(),
			Username: user.Username,
		}
		if err := internalRedis.StoreUser(rdb, &newUser); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonResponse)
	}
}

func deleteUserHandler(rdb *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		if err := internalRedis.DeleteUser(rdb, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func FileHandlers(minioClient *minio.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			downloadFileHandler(minioClient)(w, r)
		case http.MethodPost:
			uploadFileHandler(minioClient)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func downloadFileHandler(minioClient *minio.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bucketName := "default"
		objectName := strings.TrimPrefix(r.URL.Path, "/files/")
		object := internalMinio.DownloadFile(minioClient, bucketName, objectName)
		defer object.Close()

		if _, err := io.Copy(w, object); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func uploadFileHandler(minioClient *minio.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bucketName := "default"

		reader, fileHeaders, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer reader.Close()

		internalMinio.UploadFile(minioClient, bucketName, fileHeaders.Filename, reader, fileHeaders.Size, "text/plain")

		w.Header().Set("Location", fmt.Sprintf("/files/%s", fileHeaders.Filename))
		w.WriteHeader(http.StatusCreated)
	}
}
