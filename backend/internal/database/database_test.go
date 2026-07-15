package database

import (
	"context"
	"os"
	"testing"
)

func getTestDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	return dsn
}

func getTestRedisURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}
	return url
}

func TestNewPool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, err := NewPool(context.Background(), getTestDSN(t))
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}
	defer pool.Close()

	if stats := pool.Stat(); stats.TotalConns() == 0 {
		t.Error("expected at least one connection in pool")
	}
}

func TestNewRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, err := NewRedis(getTestRedisURL(t))
	if err != nil {
		t.Fatalf("NewRedis failed: %v", err)
	}
	defer client.Close()
}
