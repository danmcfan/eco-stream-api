package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go"

	internalMinio "github.com/danmcfan/eco-stream/internal/minio"
	"github.com/danmcfan/eco-stream/internal/models"
	"github.com/danmcfan/eco-stream/internal/postgres"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("{ handler: HEALTH, addr: %s }", r.RemoteAddr)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "All systems operational!")
}

func UserHandlers(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			id := strings.TrimPrefix(r.URL.Path, "/users/")
			if id == "" {
				listUsersHandler(db)(w, r)
			} else {
				retrieveUserHandler(db)(w, r)
			}
		case http.MethodPut:
			updateUserHandler(db)(w, r)
		case http.MethodPost:
			createUserHandler(db)(w, r)
		case http.MethodDelete:
			deleteUserHandler(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listUsersHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: LIST_USERS, addr: %s }", r.RemoteAddr)
		users, err := postgres.ListUsers(db)
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

func retrieveUserHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: RETRIEVE_USER, addr: %s }", r.RemoteAddr)
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		user, err := postgres.RetrieveUser(db, id)
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

func createUserHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: CREATE_USER, addr: %s }", r.RemoteAddr)
		var user models.CreateUser
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newUser := models.User{
			ID:       uuid.New().String(),
			Username: user.Username,
			IsActive: true,
		}
		if err := postgres.StoreUser(db, &newUser); err != nil {
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

func updateUserHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: UPDATE_USER, addr: %s }", r.RemoteAddr)
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		var updateUser models.UpdateUser
		if err := json.NewDecoder(r.Body).Decode(&updateUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user := models.User{
			ID:       id,
			Username: updateUser.Username,
			IsActive: updateUser.IsActive,
		}
		if err := postgres.UpdateUser(db, &user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonResponse)
	}
}

func deleteUserHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: DELETE_USER, addr: %s }", r.RemoteAddr)
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		if err := postgres.DeleteUser(db, id); err != nil {
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
		log.Printf("{ handler: DOWNLOAD_FILE, addr: %s }", r.RemoteAddr)
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
		log.Printf("{ handler: UPLOAD_FILE, addr: %s }", r.RemoteAddr)
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
