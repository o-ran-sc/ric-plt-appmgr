package sdlgoredis

import (
	"os"

	"github.com/go-redis/redis"
)

type DB struct {
	client *redis.Client
}

func Create() *DB {
	hostname := os.Getenv("DBAAS_SERVICE_HOST")
	if hostname == "" {
		hostname = "localhost"
	}
	port := os.Getenv("DBAAS_SERVICE_PORT")
	if port == "" {
		port = "6379"
	}
	redisAddress := hostname + ":" + port
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 20,
	})

	db := DB{
		client: client,
	}

	return &db
}

func (db *DB) Close() error {
	return db.Close()
}

func (db *DB) MSet(pairs ...interface{}) error {
	return db.client.MSet(pairs...).Err()
}

func (db *DB) MGet(keys []string) ([]interface{}, error) {
	val, err := db.client.MGet(keys...).Result()
	return val, err
}

func (db *DB) Del(keys []string) error {
	_, err := db.client.Del(keys...).Result()
	return err
}

func (db *DB) Keys(pattern string) ([]string, error) {
	val, err := db.client.Keys(pattern).Result()
	return val, err
}
