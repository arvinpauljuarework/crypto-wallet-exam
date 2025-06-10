package db

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// if system become larger, meaning need to load static data, etc, better to use caching
func Redis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout:  5 * time.Second,
	})
	return client
}
