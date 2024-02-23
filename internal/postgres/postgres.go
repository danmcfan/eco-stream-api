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

func RetrieveUserByUsername(db *sql.DB, username string) (*models.User, error) {
	query := "SELECT id, username, password, isActive FROM users WHERE username = $1"
	row := db.QueryRow(query, username)

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.IsActive); err != nil {
		log.Fatalf("Failed to retrieve user: %v", err)
		return nil, err
	}

	return &user, nil
}

func ListItems(db *sql.DB, username string) ([]models.Item, error) {
	query := `
        SELECT items.id, items.name, items.count, items.userId
        FROM items
        JOIN users
            ON items.userId = users.id
        WHERE users.username = $1
        ORDER BY items.id ASC
    `

	rows, err := db.Query(query, username)
	if err != nil {
		log.Fatalf("Failed to retrieve users: %v", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Count, &item.UserID); err != nil {
			log.Fatalf("Failed to retrieve users: %v", err)
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Failed to retrieve users: %v", err)
		return nil, err
	}

	return items, nil
}

func StoreItem(db *sql.DB, item *models.Item) error {
	_, err := db.Exec("INSERT INTO items (id, name, count, userId) VALUES ($1, $2, $3, $4)", item.ID, item.Name, item.Count, item.UserID)
	return err
}

func UpdateItem(db *sql.DB, item *models.Item) error {
	_, err := db.Exec("UPDATE items SET name = $1, count = $2 WHERE id = $3", item.Name, item.Count, item.ID)
	return err
}

func DeleteItem(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM items WHERE id = $1", id)
	return err
}
