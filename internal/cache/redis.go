package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is used to explicitly state the key does not exist in the cache.
var ErrNotFound = errors.New("not found")

// Redis wraps a open Redis connection.
type Redis struct {
	conn *redis.Client
}

// Open opens this Redis connection use the options provided.
func (r *Redis) Open(host, port, username, password string) {
	r.conn = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Username: username,
		Password: password,
		DB:       0,
	})
}

// Close closes this Redis connection.
func (r *Redis) Close() error {
	return r.conn.Close()
}

// Ping pings the Redis cache. Open must be called before calling this function.
func (r *Redis) Ping() error {
	if _, err := r.conn.Ping(context.Background()).Result(); err != nil {
		return err
	}

	return nil
}

// Set sets the key-value pair in the Redis cache. Open must be called before calling this function.
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return set(r.conn, ctx, SetTxParams{
		Key: key,
		Val: value,
		Exp: expiration,
	})
}

type SetTxParams struct {
	Key string
	Val interface{}
	Exp time.Duration
}

// SetTx sets all the key-value pairs in txs in the Redis cache. All sets must succeed or none will.
// Open must be called before calling this function.
func (r *Redis) SetTx(ctx context.Context, txs ...SetTxParams) error {
	pipe := r.conn.TxPipeline()

	err := set(pipe, ctx, txs...)
	if err != nil {
		return err
	}

	_, err = pipe.Exec(ctx)
	return err
}

type Setter interface {
	Set(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd
}

func set(s Setter, ctx context.Context, txs ...SetTxParams) error {
	for _, tx := range txs {
		err := s.Set(ctx, tx.Key, tx.Val, tx.Exp).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// Get gets the value for the specified key in the Redis cache. Open must be called before calling this function.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	res, err := r.conn.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNotFound
		}

		return "", err
	}

	return res, nil
}

// Del will delete all the keys from the Redis cache. All the deletes will succeed or none will.
// Open must be called before calling this function.
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.conn.Del(ctx, keys...).Err()
}
