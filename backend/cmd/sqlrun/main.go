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
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/sqlrun <query>")
		os.Exit(1)
	}

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fieldDesc := rows.FieldDescriptions()
	cols := make([]string, len(fieldDesc))
	colW := make([]int, len(fieldDesc))
	for i, fd := range fieldDesc {
		cols[i] = string(fd.Name)
		if len(cols[i]) > colW[i] {
			colW[i] = len(cols[i])
		}
	}
	fmt.Println(strings.Join(cols, " | "))

	vals := make([]interface{}, len(cols))
	for i := range vals {
		var s string
		vals[i] = &s
	}
	for rows.Next() {
		rowVals := make([]interface{}, len(cols))
		for i := range rowVals {
			var s string
			rowVals[i] = &s
		}
		rows.Scan(rowVals...)
		parts := make([]string, len(cols))
		for i, v := range rowVals {
			parts[i] = *(v.(*string))
		}
		fmt.Println(strings.Join(parts, " | "))
	}
}
