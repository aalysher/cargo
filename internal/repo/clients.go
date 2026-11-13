package repo

import (
	"context"
	"time"

	"cargo/internal/auth"
)

// Client — запись таблицы clients.
type Client struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	PhotoURL   string    `json:"photo_url"`
	Phone      string    `json:"phone"`
	ClientCode string    `json:"client_code"`
	CreatedAt  time.Time `json:"created_at"`
}

// clientCols — колонки для выборки клиента (NULL → пустая строка).
const clientCols = `id, telegram_id, coalesce(username, '') username, first_name, last_name,
	coalesce(photo_url, '') photo_url, coalesce(phone, '') phone, client_code, created_at`

// Clients — доступ к clients.
type Clients struct{ db DBTX }

// NewClients конструирует репозиторий.
func NewClients(db DBTX) *Clients { return &Clients{db: db} }

// UpsertTelegram создаёт или обновляет клиента по telegram_id; код клиента задаётся снаружи.
func (c *Clients) UpsertTelegram(ctx context.Context, tg auth.TelegramUser, clientCode string) (Client, error) {
	row := c.db.QueryRow(ctx, `
		INSERT INTO clients (telegram_id, username, first_name, last_name, photo_url, client_code)
		VALUES ($1, nullif($2, ''), $3, $4, $5, $6)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username   = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name  = EXCLUDED.last_name,
			photo_url  = EXCLUDED.photo_url,
			updated_at = now()
		RETURNING `+clientCols,
		tg.ID, tg.Username, tg.FirstName, tg.LastName, tg.PhotoURL, clientCode)

	var cl Client
	err := row.Scan(&cl.ID, &cl.TelegramID, &cl.Username, &cl.FirstName, &cl.LastName,
		&cl.PhotoURL, &cl.Phone, &cl.ClientCode, &cl.CreatedAt)
	return cl, err
}

// GetByID возвращает клиента по id.
func (c *Clients) GetByID(ctx context.Context, id int64) (Client, error) {
	row := c.db.QueryRow(ctx, `SELECT `+clientCols+` FROM clients WHERE id = $1`, id)
	var cl Client
	err := row.Scan(&cl.ID, &cl.TelegramID, &cl.Username, &cl.FirstName, &cl.LastName,
		&cl.PhotoURL, &cl.Phone, &cl.ClientCode, &cl.CreatedAt)
	return cl, err
}

// UpdatePhone сохраняет телефон клиента.
func (c *Clients) UpdatePhone(ctx context.Context, id int64, phone string) error {
	_, err := c.db.Exec(ctx,
		`UPDATE clients SET phone = nullif($2, ''), updated_at = now() WHERE id = $1`, id, phone)
	return err
}
