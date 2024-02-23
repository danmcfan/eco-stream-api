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

	"github.com/danmcfan/eco-stream/internal/jwt"
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("{ handler: LOGIN, addr: %s }", r.RemoteAddr)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var user models.LoginUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Username != "admin" || user.Password != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenString, err := jwt.CreateToken(user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(models.Token{Token: tokenString})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("{ handler: AUTHENTICATE, addr: %s }", r.RemoteAddr)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, err := jwt.AuthenticateUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ItemHandlers(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listItemsHandler(db)(w, r)
		case http.MethodPut:
			updateItemHandler(db)(w, r)
		case http.MethodPost:
			createItemHandler(db)(w, r)
		case http.MethodDelete:
			deleteItemHandler(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listItemsHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: LIST_ITEMS, addr: %s }", r.RemoteAddr)
		username, err := jwt.AuthenticateUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		items, err := postgres.ListItems(db, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}
}

func createItemHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: CREATE_ITEM, addr: %s }", r.RemoteAddr)

		username, err := jwt.AuthenticateUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := postgres.RetrieveUserByUsername(db, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var item models.CreateItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newItem := models.Item{
			ID:     uuid.New().String(),
			Name:   item.Name,
			Count:  0,
			UserID: user.ID,
		}
		if err := postgres.StoreItem(db, &newItem); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(newItem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonResponse)
	}
}

func updateItemHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: UPDATE_ITEM, addr: %s }", r.RemoteAddr)

		id := strings.TrimPrefix(r.URL.Path, "/items/")
		var updateItem models.UpdateItem
		if err := json.NewDecoder(r.Body).Decode(&updateItem); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := jwt.AuthenticateUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		item := models.Item{
			ID:     id,
			Name:   updateItem.Name,
			Count:  updateItem.Count,
			UserID: "",
		}
		if err := postgres.UpdateItem(db, &item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonResponse)
	}
}

func deleteItemHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("{ handler: DELETE_USER, addr: %s }", r.RemoteAddr)

		id := strings.TrimPrefix(r.URL.Path, "/items/")
		_, err := jwt.AuthenticateUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if err := postgres.DeleteItem(db, id); err != nil {
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
