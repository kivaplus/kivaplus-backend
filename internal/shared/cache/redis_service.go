package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService handles Redis operations for caching
type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisService creates a new Redis service
func NewRedisService() *RedisService {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379" // Default for local development
	}

	// Parse Redis URL for AWS ElastiCache or local Redis
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback to simple configuration
		opts = &redis.Options{
			Addr:     redisURL,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		}
	}

	// Configure for AWS ElastiCache
	opts.PoolSize = 10
	opts.MinIdleConns = 5
	opts.MaxRetries = 3
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	client := redis.NewClient(opts)

	return &RedisService{
		client: client,
		ctx:    context.Background(),
	}
}

// Set stores a value in Redis with expiration
func (r *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(r.ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value from Redis
func (r *RedisService) Get(key string, dest interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheNotFound
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from Redis
func (r *RedisService) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisService) Exists(key string) (bool, error) {
	count, err := r.client.Exists(r.ctx, key).Result()
	return count > 0, err
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (r *RedisService) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.SetNX(r.ctx, key, jsonValue, expiration).Result()
}

// Increment atomically increments a counter
func (r *RedisService) Increment(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// SetExpiration sets expiration for an existing key
func (r *RedisService) SetExpiration(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// GetMultiple retrieves multiple keys at once
func (r *RedisService) GetMultiple(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	values, err := r.client.MGet(r.ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple values: %w", err)
	}

	result := make(map[string]string)
	for i, key := range keys {
		if values[i] != nil {
			result[key] = values[i].(string)
		}
	}

	return result, nil
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Ping tests the Redis connection
func (r *RedisService) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// Cache key generators
func (r *RedisService) UserPermissionsKey(userID string) string {
	return fmt.Sprintf("permissions:user:%s", userID)
}

func (r *RedisService) UserTokenVersionKey(userID string) string {
	return fmt.Sprintf("token_version:user:%s", userID)
}

func (r *RedisService) UserProfileKey(userID string) string {
	return fmt.Sprintf("profile:user:%s", userID)
}

func (r *RedisService) CondominiumKey(condominiumID int64) string {
	return fmt.Sprintf("condominium:%d", condominiumID)
}

// Error definitions
var (
	ErrCacheNotFound = fmt.Errorf("cache: key not found")
)
