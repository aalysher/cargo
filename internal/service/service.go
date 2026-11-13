// Package service: бизнес-логика приложения (правила статусов, цена, события, outbox).
package service

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"cargo/internal/config"
	"cargo/internal/repo"
)

// Service — фасад над репозиториями; держит пул для транзакций.
type Service struct {
	cfg  *config.Config
	pool *pgxpool.Pool

	Clients       *repo.Clients
	Orders        *repo.Orders
	Events        *repo.Events
	Notifications *repo.Notifications
	Tariffs       *repo.Tariffs
	Settings      *repo.Settings
	Admins        *repo.Admins
	AdminOrders   *repo.AdminOrders
}

// New конструирует сервис.
func New(pool *pgxpool.Pool, cfg *config.Config) *Service {
	return &Service{
		cfg:           cfg,
		pool:          pool,
		Clients:       repo.NewClients(pool),
		Orders:        repo.NewOrders(pool),
		Events:        repo.NewEvents(pool),
		Notifications: repo.NewNotifications(pool),
		Tariffs:       repo.NewTariffs(pool),
		Settings:      repo.NewSettings(pool),
		Admins:        repo.NewAdmins(pool),
		AdminOrders:   repo.NewAdminOrders(pool),
	}
}

// Config — доступ к конфигурации для handler'ов.
func (s *Service) Config() *config.Config { return s.cfg }
