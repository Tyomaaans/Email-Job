package redis

import (
    "context"
    "log"

    "github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       0,
    })

    if err := client.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("redis: failed to connect %v!", err)
    }
    
    log.Print("redis: connection established!")
    
    return client
}