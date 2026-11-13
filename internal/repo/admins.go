package repo

import (
	"context"
	"time"
)

// AdminUser — оператор/админ (admin_users).
type AdminUser struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// adminCols — колонки для выборки оператора.
const adminCols = `id, username, password_hash, display_name, role::text, is_active, created_at`

// Admins — доступ к admin_users.
type Admins struct{ db DBTX }

// NewAdmins конструирует репозиторий.
func NewAdmins(db DBTX) *Admins { return &Admins{db: db} }

// GetByUsername — оператор по логину.
func (a *Admins) GetByUsername(ctx context.Context, username string) (AdminUser, error) {
	row := a.db.QueryRow(ctx,
		`SELECT `+adminCols+` FROM admin_users WHERE username = $1`, username)
	var u AdminUser
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.Role, &u.IsActive, &u.CreatedAt)
	return u, err
}
