package RedisClient

import (
	"github.com/redis/go-redis/v9"
)

const (
	defaultAddr     = "localhost:6379"
	defaultPassword = ""
	defaultDB       = 0
)

func NewRedisClient(addr, password *string, db *int) *redis.Client {
	finalAddr := defaultAddr
	if addr != nil {
		finalAddr = *addr
	}

	finalPassword := defaultPassword
	if password != nil {
		finalPassword = *password
	}

	finalDB := defaultDB
	if db != nil {
		finalDB = *db
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     finalAddr,
		Password: finalPassword,
		DB:       finalDB,
	})
	return rdb
}
