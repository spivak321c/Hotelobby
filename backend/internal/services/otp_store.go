package services

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOTPStore struct {
	client *redis.Client
}

func NewRedisOTPStore(client *redis.Client) *RedisOTPStore {
	return &RedisOTPStore{client: client}
}

func (s *RedisOTPStore) Set(ctx context.Context, key, otp string, ttl time.Duration) error {
	return s.client.Set(ctx, key, otp, ttl).Err()
}

func (s *RedisOTPStore) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *RedisOTPStore) Del(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

type MemoryOTPStore struct {
	mu   sync.RWMutex
	vals map[string]mEntry
}

type mEntry struct {
	otp    string
	expiry time.Time
}

func NewMemoryOTPStore() *MemoryOTPStore {
	s := &MemoryOTPStore{vals: make(map[string]mEntry)}
	go s.cleanupLoop()
	return s
}

func (s *MemoryOTPStore) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.vals {
			if now.After(v.expiry) {
				delete(s.vals, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *MemoryOTPStore) Set(ctx context.Context, key, otp string, ttl time.Duration) error {
	s.mu.Lock()
	s.vals[key] = mEntry{otp: otp, expiry: time.Now().Add(ttl)}
	s.mu.Unlock()
	return nil
}

func (s *MemoryOTPStore) Get(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	e, ok := s.vals[key]
	s.mu.RUnlock()
	if !ok || time.Now().After(e.expiry) {
		return "", redis.Nil
	}
	return e.otp, nil
}

func (s *MemoryOTPStore) Del(ctx context.Context, key string) error {
	s.mu.Lock()
	delete(s.vals, key)
	s.mu.Unlock()
	return nil
}
