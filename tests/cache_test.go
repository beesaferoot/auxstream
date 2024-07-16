package tests

import (
	"auxstream/cache"
	"auxstream/db"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetKey(t *testing.T) {
	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)

	object := &db.Artist{Name: "Test", Id: 1, CreatedAt: time.Now()}
	err := r.Get("artist-1", object)

	require.Error(t, err, redis.Nil)

	err = r.Set("artist-1", object, 1*time.Millisecond)

	require.NoError(t, err)

	object1 := &db.Artist{}
	err = r.Get("artist-1", object1)
	require.NoError(t, err)

	require.Equal(t, object.Name, object1.Name)

}

func TestSetKey(t *testing.T) {
	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)

	object := &db.Artist{Name: "Test", Id: 1, CreatedAt: time.Now()}

	err := r.Set("artist-1", object, 1*time.Millisecond)
	require.NoError(t, err)
	object1 := &db.Artist{}
	err = r.Get("artist-1", object1)
	require.NoError(t, err)
	require.Equal(t, object.Name, object1.Name)
}

func TestDeleteKey(t *testing.T) {

	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	r := cache.NewRedis(opts)
	object := &db.Artist{Name: "Test", Id: 1, CreatedAt: time.Now()}
	err := r.Set("artist-1", object, 1*time.Millisecond)
	require.NoError(t, err)

	err = r.Del("artist-1")

	require.NoError(t, err)

	object1 := &db.Artist{}

	err = r.Get("artist-1", object1)

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
