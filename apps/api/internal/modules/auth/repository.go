package auth

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct{ DB *sql.DB }

type userWithHash struct {
	User
	PasswordHash string
	IsActive     bool
}

func (r Repository) CreateUser(ctx context.Context, id, email, displayName, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO users(id,email,display_name,password_hash) VALUES(?,?,?,?)`, id, email, displayName, passwordHash)
	return err
}

func (r Repository) UserByEmail(ctx context.Context, email string) (userWithHash, error) {
	var u userWithHash
	var active int
	err := r.DB.QueryRowContext(ctx, `SELECT id,email,display_name,password_hash,is_active FROM users WHERE email=?`, email).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &active)
	u.IsActive = active == 1
	return u, err
}

func (r Repository) UserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := r.DB.QueryRowContext(ctx, `SELECT id,email,display_name FROM users WHERE id=? AND is_active=1`, id).
		Scan(&u.ID, &u.Email, &u.DisplayName)
	return u, err
}

func (r Repository) CreateSession(ctx context.Context, id, userID, tokenHash, expiresAt, userAgent string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO sessions(id,user_id,token_hash,expires_at,user_agent) VALUES(?,?,?,?,?)`, id, userID, tokenHash, expiresAt, userAgent)
	return err
}

func (r Repository) UserIDBySessionHash(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.DB.QueryRowContext(ctx, `SELECT user_id FROM sessions WHERE token_hash=? AND expires_at > CURRENT_TIMESTAMP`, tokenHash).Scan(&userID)
	if err == nil {
		_, _ = r.DB.ExecContext(ctx, `UPDATE sessions SET last_seen_at=CURRENT_TIMESTAMP WHERE token_hash=?`, tokenHash)
	}
	return userID, err
}

func (r Repository) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash=?`, tokenHash)
	return err
}

func (r Repository) CompanyForUser(ctx context.Context, userID string) (*CompanyContext, error) {
	var c CompanyContext
	err := r.DB.QueryRowContext(ctx, `
		SELECT c.id,c.name,c.slug,c.timezone,c.currency,m.role
		FROM company_members m JOIN companies c ON c.id=m.company_id
		WHERE m.user_id=? ORDER BY m.created_at ASC LIMIT 1`, userID).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Timezone, &c.Currency, &c.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
