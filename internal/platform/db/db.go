package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"pluto-backend/internal/platform/config"
)

func warmUpDB(db *sql.DB, n int) error {
	var wg sync.WaitGroup
	errCh := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := db.Conn(context.Background())
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close() // Вернуть в пул
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func NewDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := warmUpDB(db, 5); err != nil {
		_ = db.Close()
		return nil, err
	}

	stats := db.Stats()
	fmt.Printf("OpenConnections: %d\n", stats.OpenConnections)
	fmt.Printf("IdleConnections: %d\n", stats.Idle)
	fmt.Printf("InUseConnections: %d\n", stats.InUse)

	return db, nil
}
