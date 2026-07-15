package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var globalPool *pgxpool.Pool

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" || testing.Short() {
		t.Skip("DATABASE_URL not set or -short")
	}
	if globalPool == nil {
		var err error
		globalPool, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			t.Fatalf("test pool: %v", err)
		}
	}
	return globalPool
}

func TestMain(m *testing.M) {
	code := m.Run()
	if globalPool != nil {
		globalPool.Close()
	}
	os.Exit(code)
}
