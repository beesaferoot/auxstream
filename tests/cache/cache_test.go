package tests

import (
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*cache.Redis, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	opts := &redis.Options{Addr: mr.Addr()}
	r := cache.NewRedis(opts)
	return r, mr
}

func TestGetKey(t *testing.T) {
	r, _ := setupTestRedis(t)
	m1 := &db.Artist{Name: "Test", ID: uuid.New(), CreatedAt: time.Now()}
	object := &cache.Cacheable[db.Artist]{Value: m1}
	err := r.Get("artist-1", object)

	require.Error(t, err, redis.Nil)

	err = r.Set("artist-1", object, 1*time.Millisecond)

	require.NoError(t, err)

	m2 := &db.Artist{}
	object1 := &cache.Cacheable[db.Artist]{Value: m2}
	err = r.Get("artist-1", object1)
	require.NoError(t, err)

	require.Equal(t, m1.Name, m2.Name)

}

func TestSetKey(t *testing.T) {
	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)

	m1 := &db.Artist{Name: "Test", ID: uuid.New(), CreatedAt: time.Now()}

	err := r.Set("artist-1", &cache.Cacheable[db.Artist]{Value: m1}, 1*time.Millisecond)
	require.NoError(t, err)
	m2 := &db.Artist{}
	err = r.Get("artist-1", &cache.Cacheable[db.Artist]{Value: m2})
	require.NoError(t, err)
	require.Equal(t, "Test", m2.Name)
}

func TestDeleteKey(t *testing.T) {
	r, _ := setupTestRedis(t)
	m1 := &db.Artist{Name: "Test", ID: uuid.New(), CreatedAt: time.Now()}
	err := r.Set("artist-1", &cache.Cacheable[db.Artist]{Value: m1}, 1*time.Millisecond)
	require.NoError(t, err)

	err = r.Del("artist-1")

	require.NoError(t, err)

	m2 := &db.Artist{}

	err = r.Get("artist-1", &cache.Cacheable[db.Artist]{Value: m2})

	require.Error(t, err, redis.Nil)
}

func TestSetStringKey(t *testing.T) {
	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)

	_, err := r.GetString("test-val")

	require.Error(t, err, redis.Nil)

	err = r.SetString("test-val", "value", 1*time.Millisecond)

	require.NoError(t, err)

	val, err := r.GetString("test-val")

	require.NoError(t, err)

	require.Equal(t, "value", val)
}

func TestGetMissingKey(t *testing.T) {
	r, _ := setupTestRedis(t)

	m := &db.Artist{}
	err := r.Get("missing", &cache.Cacheable[db.Artist]{Value: m})
	require.Error(t, err)
	require.Equal(t, redis.Nil, err)
}

func TestSetStringAndGetString(t *testing.T) {
	r, _ := setupTestRedis(t)

	err := r.SetString("hello", "world", time.Second)
	require.NoError(t, err)

	val, err := r.GetString("hello")
	require.NoError(t, err)
	require.Equal(t, "world", val)
}

func TestExistsAndTTL(t *testing.T) {
	r, _ := setupTestRedis(t)

	err := r.SetString("ttl-key", "exists", 500*time.Millisecond)
	require.NoError(t, err)

	ctx := context.Background()
	exists, err := r.Exists(ctx, "ttl-key")
	require.NoError(t, err)
	require.True(t, exists)

	ttl, err := r.TTL(ctx, "ttl-key")
	require.NoError(t, err)
	require.True(t, ttl > 0)

	time.Sleep(600 * time.Millisecond)
	exists, _ = r.Exists(ctx, "ttl-key")
	require.False(t, exists)
}

func TestExpireKey(t *testing.T) {
	r, _ := setupTestRedis(t)
	ctx := context.Background()

	err := r.SetString("exp-key", "hello", 0)
	require.NoError(t, err)

	err = r.Expire(ctx, "exp-key", 200*time.Millisecond)
	require.NoError(t, err)

	time.Sleep(250 * time.Millisecond)
	exists, _ := r.Exists(ctx, "exp-key")
	require.False(t, exists)
}

func TestSetOperations(t *testing.T) {
	r, _ := setupTestRedis(t)
	ctx := context.Background()

	err := r.SAdd(ctx, "user:tokens", "t1", "t2", "t3")
	require.NoError(t, err)

	members, err := r.SMembers(ctx, "user:tokens")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"t1", "t2", "t3"}, members)

	err = r.SRem(ctx, "user:tokens", "t1", "t3")
	require.NoError(t, err)

	members, err = r.SMembers(ctx, "user:tokens")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"t2"}, members)
}

func TestIncrDecr(t *testing.T) {
	r, _ := setupTestRedis(t)

	ctx := context.Background()

	v, err := r.Incr(ctx, "counter")
	require.NoError(t, err)
	require.Equal(t, int64(1), v)

	v, err = r.Incr(ctx, "counter")
	require.NoError(t, err)
	require.Equal(t, int64(2), v)

	v, err = r.Decr(ctx, "counter")
	require.NoError(t, err)
	require.Equal(t, int64(1), v)
}
