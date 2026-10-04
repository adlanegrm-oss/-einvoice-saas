package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	FullName     string    `json:"full_name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) CreateUser(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, role, full_name, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query, u.ID, u.TenantID, u.Email, u.PasswordHash, u.Role, u.FullName, u.IsActive, now, now)
	if err != nil {
		return fmt.Errorf("user_store: failed to create user: %w", err)
	}
	return nil
}

func (s *UserStore) GetByEmail(ctx context.Context, tenantID, email string) (*User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, full_name, is_active, created_at, updated_at
		FROM users
		WHERE tenant_id = ? AND email = ?
	`
	row := s.db.QueryRowContext(ctx, query, tenantID, email)

	var u User
	err := row.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}
