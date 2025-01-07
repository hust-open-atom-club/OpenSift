package storage

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func InitRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return rdb, nil
}

func SetKeyValue(rdb *redis.Client, key, value string) error {
	err := rdb.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("could not set key '%s': %v", key, err)
	}
	return nil
}

func GetKeyValue(rdb *redis.Client, key string) (string, error) {
	val, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key '%s' does not exist", key)
		}
		return "", fmt.Errorf("could not get key '%s': %v", key, err)
	}
	return val, nil
}

func PersistData(rdb *redis.Client) error {
	err := rdb.BgSave(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("could not trigger RDB save: %v", err)
	}
	return nil
}
