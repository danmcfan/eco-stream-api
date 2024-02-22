package postgres

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/danmcfan/eco-stream/internal/models"
)

func CreatePostgresClient() *sql.DB {
	postgresURL := "postgres://localhost:5432"
	if val, ok := os.LookupEnv("POSTGRES_URL"); ok {
		postgresURL = val
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	return db
}

func ListUsers(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query("SELECT id, username, isActive FROM users ORDER BY id ASC")
	if err != nil {
		log.Fatalf("Failed to retrieve users: %v", err)
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsActive); err != nil {
			log.Fatalf("Failed to retrieve users: %v", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Failed to retrieve users: %v", err)
		return nil, err
	}

	return users, nil
}

func RetrieveUser(db *sql.DB, id string) (*models.User, error) {
	var user models.User
	err := db.QueryRow("SELECT id, username, isActive FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username, &user.IsActive)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Fatalf("Failed to retrieve user: %v", err)
		return nil, err
	}

	return &user, nil
}

func StoreUser(db *sql.DB, user *models.User) error {
	_, err := db.Exec("INSERT INTO users (id, username, isActive) VALUES ($1, $2, $3)", user.ID, user.Username, user.IsActive)
	return err
}

func UpdateUser(db *sql.DB, user *models.User) error {
	_, err := db.Exec("UPDATE users SET username = $1, isActive = $2 WHERE id = $3", user.Username, user.IsActive, user.ID)
	return err
}

func DeleteUser(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}
