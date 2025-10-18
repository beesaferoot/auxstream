package tests

import (
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestGetKey(t *testing.T) {
	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)

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

	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)
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
