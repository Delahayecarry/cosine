package database

import (
	"database/sql"
	"time"
)

// LinuxDoUser represents a user from LinuxDo OAuth
type LinuxDoUser struct {
	ID         int64     `json:"id"`
	LinuxDoID  int       `json:"linuxdo_id"`
	Username   string    `json:"username"`
	Name       string    `json:"name"`
	TrustLevel int       `json:"trust_level"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateOrUpdateLinuxDoUser creates a new user or updates existing one based on linuxdo_id
func CreateOrUpdateLinuxDoUser(linuxDoID int, username, name string, trustLevel int, active bool) (*LinuxDoUser, error) {
	var user LinuxDoUser

	// Try to find existing user
	err := db.QueryRow(`
		SELECT id, linuxdo_id, username, name, trust_level, active, created_at, updated_at
		FROM linuxdo_user
		WHERE linuxdo_id = $1
	`, linuxDoID).Scan(
		&user.ID, &user.LinuxDoID, &user.Username, &user.Name,
		&user.TrustLevel, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new user
		err = db.QueryRow(`
			INSERT INTO linuxdo_user (linuxdo_id, username, name, trust_level, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			RETURNING id, linuxdo_id, username, name, trust_level, active, created_at, updated_at
		`, linuxDoID, username, name, trustLevel, active).Scan(
			&user.ID, &user.LinuxDoID, &user.Username, &user.Name,
			&user.TrustLevel, &user.Active, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	if err != nil {
		return nil, err
	}

	// Update existing user
	err = db.QueryRow(`
		UPDATE linuxdo_user
		SET username = $2, name = $3, trust_level = $4, active = $5, updated_at = NOW()
		WHERE linuxdo_id = $1
		RETURNING id, linuxdo_id, username, name, trust_level, active, created_at, updated_at
	`, linuxDoID, username, name, trustLevel, active).Scan(
		&user.ID, &user.LinuxDoID, &user.Username, &user.Name,
		&user.TrustLevel, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetLinuxDoUserByLinuxDoID retrieves a user by their LinuxDo ID
func GetLinuxDoUserByLinuxDoID(linuxDoID int) (*LinuxDoUser, error) {
	var user LinuxDoUser
	err := db.QueryRow(`
		SELECT id, linuxdo_id, username, name, trust_level, active, created_at, updated_at
		FROM linuxdo_user
		WHERE linuxdo_id = $1
	`, linuxDoID).Scan(
		&user.ID, &user.LinuxDoID, &user.Username, &user.Name,
		&user.TrustLevel, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
