package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a key-value store with expiry, set, and counter operations.
// The any-typed and string-typed accessors are not interchangeable: Set/Get
// round-trip values as JSON, whereas SetString/GetString store the raw bytes,
// so a key written by one pair cannot be read back by the other.
type Cache interface {
	// Set stores value as JSON under key, expiring after exp (zero means no expiry).
	Set(key string, value any, exp time.Duration) error
	SetString(key string, value string, exp time.Duration) error
	// Get reads the JSON at key and unmarshals it into value, which must be a pointer.
	Get(key string, value any) error
	GetString(key string) (string, error)
	Del(key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, exp time.Duration) error
	// TTL reports the remaining lifetime of key at millisecond precision.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Collection helpers
	SAdd(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...string) error

	// Numeric helpers
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
}

// Redis is the go-redis backed Cache. Its methods that take no context use a
// background context internally.
type Redis struct {
	client *redis.Client
}

// NewRedis returns a Redis using a client built from opts; it does not dial until
// the first operation, so a bad address surfaces as an error there, not here.
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

func (r *Redis) Get(key string, value any) error {
	result := r.client.Get(context.Background(), key)
	if err := result.Err(); err != nil {
		return err
	}

	vBytes, err := result.Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(vBytes, value)
}

func (r *Redis) GetString(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
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
	// PTTL reports millisecond precision; TTL rounds to whole seconds and would
	// report 0 for any sub-second remaining lifetime.
	return r.client.PTTL(ctx, key).Result()
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
