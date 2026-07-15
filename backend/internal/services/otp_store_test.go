package services

import (
	"context"
	"testing"
	"time"
)

func TestMemoryOTPStore_SetGetDel(t *testing.T) {
	store := NewMemoryOTPStore()
	ctx := context.Background()

	if err := store.Set(ctx, "key1", "123456", 15*time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}

	otp, err := store.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if otp != "123456" {
		t.Fatalf("expected 123456, got %s", otp)
	}

	if err := store.Del(ctx, "key1"); err != nil {
		t.Fatalf("Del: %v", err)
	}

	otp, err = store.Get(ctx, "key1")
	if err == nil {
		t.Fatalf("expected error after delete, got otp=%s", otp)
	}
}

func TestMemoryOTPStore_Get_missing(t *testing.T) {
	store := NewMemoryOTPStore()
	ctx := context.Background()

	otp, err := store.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatalf("expected error for missing key, got otp=%s", otp)
	}
}

func TestMemoryOTPStore_Get_expired(t *testing.T) {
	store := NewMemoryOTPStore()
	ctx := context.Background()

	// Set with a very short TTL
	if err := store.Set(ctx, "expire_key", "999999", 1*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Wait for expiry
	time.Sleep(5 * time.Millisecond)

	otp, err := store.Get(ctx, "expire_key")
	if err == nil {
		t.Fatalf("expected error for expired key, got otp=%s", otp)
	}
}

func TestMemoryOTPStore_Overwrite(t *testing.T) {
	store := NewMemoryOTPStore()
	ctx := context.Background()

	store.Set(ctx, "key", "first", 15*time.Minute)
	store.Set(ctx, "key", "second", 15*time.Minute)

	otp, err := store.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if otp != "second" {
		t.Fatalf("expected second, got %s", otp)
	}
}

func TestMemoryOTPStore_Del_nonexistent(t *testing.T) {
	store := NewMemoryOTPStore()
	ctx := context.Background()

	// Deleting a non-existent key should not error
	if err := store.Del(ctx, "nonexistent"); err != nil {
		t.Fatalf("Del: %v", err)
	}
}

func TestMemoryOTPStore_TTLBoundary(t *testing.T) {
	store := NewMemoryOTPStore()
	ctx := context.Background()

	// Set with 10ms TTL
	if err := store.Set(ctx, "ttl_key", "abc", 10*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Immediately get — should work
	otp, err := store.Get(ctx, "ttl_key")
	if err != nil {
		t.Fatalf("Get immediately after Set: %v", err)
	}
	if otp != "abc" {
		t.Fatalf("expected abc, got %s", otp)
	}
}
