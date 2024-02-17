package redis

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/go-redis/redis/v8"

	"github.com/danmcfan/eco-stream/internal/models"
)

var ctx = context.Background()

func CreateRedisClient() *redis.Client {
	redisURL := "redis://localhost:6379"
	if val, ok := os.LookupEnv("REDIS_URL"); ok {
		redisURL = val
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)
	return client
}

func ListUsers(client *redis.Client) ([]models.User, error) {
	users := []models.User{}
	vals, err := client.HGetAll(ctx, "users").Result()
	if err != nil {
		return nil, err
	}
	for _, v := range vals {
		var user models.User
		if err := json.Unmarshal([]byte(v), &user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// Sort users by username
	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})

	return users, nil
}

func RetrieveUser(client *redis.Client, id string) (*models.User, error) {
	val, err := client.HGet(ctx, "users", id).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var user models.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func StoreUser(client *redis.Client, user *models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	_, err = client.HSet(ctx, "users", user.ID, data).Result()
	return err
}

func DeleteUser(client *redis.Client, id string) error {
	_, err := client.HDel(ctx, "users", id).Result()
	return err
}
