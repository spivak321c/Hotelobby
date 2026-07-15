package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := ""
	envPaths := []string{".env", "../.env", "../../.env"}
	for _, p := range envPaths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "DATABASE_URL=") {
				dsn = strings.TrimSpace(line)[13:]
				dsn = strings.Trim(dsn, `"'`)
			}
		}
		if dsn != "" {
			break
		}
	}
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL not found")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	query := strings.Join(os.Args[1:], " ")
	if query == "" {
		b, _ := io.ReadAll(os.Stdin)
		query = string(b)
	}
	query = strings.TrimSpace(query)
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/explain <query>")
		os.Exit(1)
	}

	explain := "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) " + query
	rows, err := pool.Query(context.Background(), explain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explain error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var line string
		rows.Scan(&line)
		fmt.Println(line)
	}
}
