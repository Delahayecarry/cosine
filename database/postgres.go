package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"cosine/config"
	"cosine/models"

	_ "github.com/lib/pq"
)

var (
	db      *sql.DB
	counter uint64
	mu      sync.RWMutex
)

func Init(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}

// GetNextAccount 使用 Round-Robin 获取下一个可用账户
func GetNextAccount() (*models.Account, error) {
	mu.RLock()
	defer mu.RUnlock()

	accounts, err := getActiveAccounts()
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no active accounts available")
	}

	idx := atomic.AddUint64(&counter, 1) % uint64(len(accounts))
	return &accounts[idx], nil
}

// getActiveAccounts 获取所有活跃账户
func getActiveAccounts() ([]models.Account, error) {
	rows, err := db.Query(`
		SELECT id, auth, team_id, linuxdo_id, is_active, created_at, updated_at
		FROM accounts
		WHERE is_active = true
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(
			&acc.ID, &acc.Auth, &acc.TeamID, &acc.LinuxdoID,
			&acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

// DeactivateAccount 将账户标记为不活跃
func DeactivateAccount(accountID int) error {
	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(`
		UPDATE accounts
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`, accountID)
	if err != nil {
		return fmt.Errorf("failed to deactivate account %d: %w", accountID, err)
	}

	log.Printf("Account %d has been deactivated", accountID)
	return nil
}

// GetAccountCount 获取活跃账户数量
func GetAccountCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE is_active = true`).Scan(&count)
	return count, err
}

// CreateAccount 创建新的捐赠账户
func CreateAccount(auth, teamID string, linuxdoID int) (*models.Account, error) {
	mu.Lock()
	defer mu.Unlock()

	var acc models.Account

	err := db.QueryRow(`
		INSERT INTO accounts (auth, team_id, linuxdo_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id, auth, team_id, linuxdo_id, is_active, created_at, updated_at
	`, auth, teamID, linuxdoID).Scan(
		&acc.ID, &acc.Auth, &acc.TeamID, &acc.LinuxdoID,
		&acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	log.Printf("New account donated by linuxdo_id: %d", linuxdoID)
	return &acc, nil
}

// GetAccountsByLinuxdoID 获取指定用户捐赠的所有账户
func GetAccountsByLinuxdoID(linuxdoID int) ([]models.Account, error) {
	rows, err := db.Query(`
		SELECT id, auth, team_id, linuxdo_id, is_active, created_at, updated_at
		FROM accounts
		WHERE linuxdo_id = $1
		ORDER BY created_at DESC
	`, linuxdoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(
			&acc.ID, &acc.Auth, &acc.TeamID, &acc.LinuxdoID,
			&acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}
