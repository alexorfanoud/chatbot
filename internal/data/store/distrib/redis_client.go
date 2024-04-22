package distrib

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"sync"
)

var (
	instance *redis.Client
	once     sync.Once
)

func GetRedisClient() *redis.Client {
	once.Do(func() {
		instance = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
		if err := redisotel.InstrumentTracing(instance); err != nil {
			panic(err)
		}
		if err := redisotel.InstrumentMetrics(instance); err != nil {
			panic(err)
		}
	})

	return instance
}
