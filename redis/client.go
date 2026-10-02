package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Init(cfg Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  4 * time.Second,
	})

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

func Ping() error {
	_, err := Client.Ping(Ctx).Result()
	return err
}

func DeleteClient(clientID string) {
	Client.Del(Ctx, "client:"+clientID)
}
