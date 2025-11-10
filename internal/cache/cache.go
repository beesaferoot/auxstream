package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// type Cacheable[T any] struct {
// 	Value *T
// }

// func (c *Cacheable[T]) MarshalBinary() ([]byte, error) {
// 	return json.Marshal(c.Value)
// }

// func (c *Cacheable[T]) UnmarshalBinary(data []byte) error {
// 	return json.Unmarshal(data, &c.Value)
// }

// type Unmarshable interface {
// 	UnmarshalBinary(data []byte) error
// }

// type Marshable interface {
// 	MarshalBinary() ([]byte, error)
// }

type Cache interface {
	Set(key string, value any, exp time.Duration) error
	SetString(key string, value string, exp time.Duration) error
	Get(key string, value any) error
	GetString(key string) (string, error)
	Del(key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, exp time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Collection helpers
	SAdd(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...string) error

	// Numeric helpers
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
}

type Redis struct {
	client *redis.Client
}

func NewRedis(opts *redis.Options) *Redis {
	return &Redis{client: redis.NewClient(opts)}
}

func (r *Redis) Set(key string, value any, exp time.Duration) error {
	encodedValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), key, encodedValue, exp).Err()
}

func (r *Redis) SetString(key string, value string, exp time.Duration) error {
	return r.client.Set(context.Background(), key, value, exp).Err()
}

func (r *Redis) Get(key string, value any) (err error) {
	result := r.client.Get(context.Background(), key)
	err = result.Err()

	if err != nil {
		return
	}

	vBytes, err := result.Bytes()

	if err != nil {
		return
	}

	err = json.Unmarshal(vBytes, value)
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

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *Redis) Expire(ctx context.Context, key string, exp time.Duration) error {
	return r.client.Expire(ctx, key, exp).Err()
}

func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *Redis) SAdd(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return errors.New("SAdd requires at least one member")
	}
	interfaceMembers := make([]any, len(members))
	for i, v := range members {
		interfaceMembers[i] = v
	}
	return r.client.SAdd(ctx, key, interfaceMembers...).Err()
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

func (r *Redis) SRem(ctx context.Context, key string, members ...string) error {
	interfaceMembers := make([]any, len(members))
	for i, v := range members {
		interfaceMembers[i] = v
	}
	return r.client.SRem(ctx, key, interfaceMembers...).Err()
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}
