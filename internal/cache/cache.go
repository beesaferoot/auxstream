package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cacheable[T any] struct {
	Value *T
}

func (c *Cacheable[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(c.Value)
}

func (c *Cacheable[T]) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &c.Value)
}

type Unmarshable interface {
	UnmarshalBinary(data []byte) error
}

type Marshable interface {
	MarshalBinary() ([]byte, error)
}

type Cache interface {
	Set(key string, value Marshable, exp time.Duration) error
	SetString(key string, value string, exp time.Duration) error
	Get(key string, value Unmarshable) error
	GetString(key string) (string, error)
	Del(key string) error
}

type Redis struct {
	client *redis.Client
}

func NewRedis(opts *redis.Options) *Redis {
	return &Redis{client: redis.NewClient(opts)}
}

func (r *Redis) Set(key string, value Marshable, exp time.Duration) error {
	return r.client.Set(context.Background(), key, value, exp).Err()
}

func (r *Redis) SetString(key string, value string, exp time.Duration) error {
	return r.client.Set(context.Background(), key, value, exp).Err()
}

func (r *Redis) Get(key string, value Unmarshable) (err error) {
	result := r.client.Get(context.Background(), key)
	err = result.Err()

	if err != nil {
		return
	}

	v, err := result.Bytes()

	if err != nil {
		return
	}

	err = value.UnmarshalBinary(v)
	return
}

func (r *Redis) GetString(key string) (string, error) {
	result := r.client.Get(context.Background(), key)

	err := result.Err()

	if err != nil {
		return "", err
	}

	return result.Result()
}

func (r *Redis) Del(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
