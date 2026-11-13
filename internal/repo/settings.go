package repo

import (
	"context"
)

// Settings — доступ к key-value настройкам (jsonb).
type Settings struct{ db DBTX }

// NewSettings конструирует репозиторий.
func NewSettings(db DBTX) *Settings { return &Settings{db: db} }

// Get возвращает значение настройки (JSON-строка).
func (s *Settings) Get(ctx context.Context, key string) (string, error) {
	var v string
	err := s.db.QueryRow(ctx, `SELECT value::text FROM settings WHERE key = $1`, key).Scan(&v)
	return v, err
}

// Put создаёт или обновляет настройку.
func (s *Settings) Put(ctx context.Context, key, value string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO settings (key, value) VALUES ($1, $2::jsonb)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`, key, value)
	return err
}
