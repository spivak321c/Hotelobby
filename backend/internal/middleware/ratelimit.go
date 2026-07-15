package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type windowEntry struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	windows  map[string]*windowEntry
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		windows: make(map[string]*windowEntry),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Middleware(c *fiber.Ctx) error {
	ip := c.IP()
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
		if ip == "" {
			ip = "unknown"
		}
	}

	rl.mu.Lock()
	now := time.Now()

	entry, exists := rl.windows[ip]
	if !exists || now.After(entry.resetAt) {
		rl.windows[ip] = &windowEntry{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		rl.mu.Unlock()
		return c.Next()
	}

	entry.count++
	if entry.count > rl.limit {
		rl.mu.Unlock()
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "rate_limited",
				"message": "too many requests, please try again later",
			},
		})
	}

	rl.mu.Unlock()
	return c.Next()
}

// Cleanup runs a periodic goroutine to purge expired entries.
// Call in a defer or use a background goroutine.
func (rl *RateLimiter) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, entry := range rl.windows {
				if now.After(entry.resetAt) {
					delete(rl.windows, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
}
