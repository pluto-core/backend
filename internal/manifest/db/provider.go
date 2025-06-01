package db

//
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"net/url"
//	"strings"
//	"sync"
//	"time"
//
//	"github.com/hashicorp/vault/api"
//	_ "github.com/jackc/pgx/v5/stdlib"
//
//	"pluto-backend/internal/manifest/config"
//)
//
//// DBProvider manages a dynamic *sql.DB pool whose credentials rotate via Vault.
//// It reads a connection_url template and dynamic credentials from Vault, builds
//// a new *sql.DB, and atomically swaps it in, with automatic refresh.
//
//type Provider struct {
//	mu          sync.RWMutex
//	cfg         *config.ManifestConfig
//	vaultClient *api.Client
//
//	db       *sql.DB
//	leaseTTL time.Duration
//	leaseID  string
//}
//
//// NewDBProvider initializes the pool with the first set of credentials.
//func NewDBProvider(cfg *config.ManifestConfig, vaultClient *api.Client) (*Provider, error) {
//	p := &Provider{cfg: cfg, vaultClient: vaultClient}
//	if err := p.refresh(); err != nil {
//		return nil, err
//	}
//	return p, nil
//}
//
//// Get returns the current *sql.DB instance.
//func (p *Provider) Get() *sql.DB {
//	p.mu.RLock()
//	defer p.mu.RUnlock()
//	return p.db
//}
//
//// StartAutoRefresh launches a background goroutine that refreshes credentials every half-lease.
//func (p *Provider) StartAutoRefresh(ctx context.Context) {
//	go func() {
//		for {
//			wait := p.leaseTTL / 2
//			if wait < time.Minute {
//				wait = time.Minute
//			}
//			select {
//			case <-time.After(wait):
//				if err := p.refresh(); err != nil {
//					fmt.Printf("db refresh error: %v\n", err)
//				}
//			case <-ctx.Done():
//				return
//			}
//		}
//	}()
//}
//
//// refresh retrieves fresh credentials and connection_url, then swaps the *sql.DB.
//func (p *Provider) refresh() error {
//	// 1) Read the DB config (connection_url template)
//	configPath := strings.Replace(p.cfg.Vault.DBPath, "creds", "config", 1)
//	configSecret, err := p.vaultClient.Logical().Read(configPath)
//	if err != nil {
//		return fmt.Errorf("vault read db config: %w", err)
//	}
//	// Extract template from "connection_details" or root
//	var tmpl string
//	if details, ok := configSecret.Data["connection_details"].(map[string]interface{}); ok {
//		if cu, ok2 := details["connection_url"].(string); ok2 {
//			tmpl = cu
//		}
//	}
//	if tmpl == "" {
//		if cu, ok := configSecret.Data["connection_url"].(string); ok {
//			tmpl = cu
//		}
//	}
//	if tmpl == "" {
//		return fmt.Errorf("connection_url missing in %s", configPath)
//	}
//
//	// 2) Read dynamic creds
//	credSecret, err := p.vaultClient.Logical().Read(p.cfg.Vault.DBPath)
//	if err != nil {
//		return fmt.Errorf("vault read db creds: %w", err)
//	}
//	var user, pass string
//	// Vault DB engine may return creds under .Data["data"] or directly under .Data
//	if dataMap, ok := credSecret.Data["data"].(map[string]interface{}); ok {
//		user = dataMap["username"].(string)
//		pass = dataMap["password"].(string)
//	} else if u, ok := credSecret.Data["username"].(string); ok {
//		user = u
//		pass = credSecret.Data["password"].(string)
//	} else {
//		return fmt.Errorf("invalid format for db creds")
//	}
//	// update lease info
//	p.leaseID = credSecret.LeaseID
//	p.leaseTTL = time.Duration(credSecret.LeaseDuration) * time.Second
//
//	// 3) Build DSN by injecting user/pass into template
//	dsn := strings.NewReplacer(
//		"{{username}}", url.QueryEscape(user),
//		"{{password}}", url.QueryEscape(pass),
//	).Replace(tmpl)
//
//	if p.cfg.Vault.AuthMethod == "token" {
//		dsn = strings.Replace(dsn, "@postgres", "@localhost", 1)
//	}
//
//	// 4) Open new DB pool
//	newDB, err := NewDB(dsn)
//	if err != nil {
//		return fmt.Errorf("open db: %w", err)
//	}
//
//	// 5) Swap pools
//	p.mu.Lock()
//	old := p.db
//	p.db = newDB
//	p.mu.Unlock()
//
//	// 6) Close old pool
//	if old != nil {
//		_ = old.Close()
//	}
//	return nil
//}
//
//// NewDB sets pool parameters and warms up connections.
//func NewDB(dsn string) (*sql.DB, error) {
//	dbPool, err := sql.Open("pgx", dsn)
//	if err != nil {
//		return nil, err
//	}
//	dbPool.SetMaxOpenConns(20)
//	dbPool.SetMaxIdleConns(5)
//	dbPool.SetConnMaxIdleTime(5 * time.Minute)
//	dbPool.SetConnMaxLifetime(30 * time.Minute)
//
//	// warm up connections
//	for i := 0; i < 5; i++ {
//		conn, err := dbPool.Conn(context.Background())
//		if err != nil {
//			dbPool.Close()
//			return nil, err
//		}
//		conn.Close()
//	}
//	return dbPool, nil
//}
