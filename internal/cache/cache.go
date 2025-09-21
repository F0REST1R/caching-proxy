package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	GetCache(ctx context.Context, key string) ([]byte, error)
	SetCache(ctx context.Context, key string, value []byte) error
	Clear(ctx context.Context) error
	Close () error
}


type RedisCache struct {
	client *redis.Client
	ctx context.Context 
}

//Функция для создания нового клиента Redis
func NewRedisClient(addr string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: "",
		DB: 0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil{
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx: ctx,
	}, nil
} 

func (r RedisCache) GetCache(ctx context.Context, key string) ([]byte, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil{
		return nil, nil
	}
	return data, err
}

func (r RedisCache) SetCache(ctx context.Context, key string, value []byte) error {
	return r.client.Set(ctx, key, value, 10*time.Minute).Err()
}

func (r RedisCache) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
} 

func (r RedisCache) Close() error {
	return r.client.Close()
}

//TTL (Time-To-Live) - это "время жизни" данных в кэше. 
//После того как TTL истекает, данные автоматически удаляются из кэша.