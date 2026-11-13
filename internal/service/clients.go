package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"cargo/internal/auth"
	"cargo/internal/domain"
	"cargo/internal/repo"
)

// clientCodeAlphabet — без похожих символов (0/O, 1/I).
const clientCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// phoneRe — формат телефона +996555123456.
var phoneRe = regexp.MustCompile(`^\+?\d{7,15}$`)

// genClientCode — личный код клиента для склада (6 символов).
func genClientCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = clientCodeAlphabet[rand.IntN(len(clientCodeAlphabet))]
	}
	return string(b)
}

// UpsertTelegramClient создаёт/обновляет клиента по initData;
// при коллизии личного кода повторяет генерацию.
func (s *Service) UpsertTelegramClient(ctx context.Context, tg auth.TelegramUser) (repo.Client, error) {
	for range 5 {
		cl, err := s.Clients.UpsertTelegram(ctx, tg, genClientCode())
		if err == nil {
			return cl, nil
		}
		if !repo.IsUniqueViolation(err) {
			return repo.Client{}, err
		}
	}
	return repo.Client{}, errors.New("не удалось сгенерировать уникальный код клиента")
}

// GetProfile — профиль клиента.
func (s *Service) GetProfile(ctx context.Context, clientID int64) (repo.Client, error) {
	cl, err := s.Clients.GetByID(ctx, clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.Client{}, domain.ErrNotFound
	}
	return cl, err
}

// UpdatePhone — телефон клиента (для связи при выдаче).
func (s *Service) UpdatePhone(ctx context.Context, clientID int64, phone string) error {
	phone = strings.TrimSpace(phone)
	if phone != "" && !phoneRe.MatchString(phone) {
		return domain.ValidationErr("телефон в формате +996555123456")
	}
	return s.Clients.UpdatePhone(ctx, clientID, phone)
}

// WarehouseAddress — адрес склада в Китае (настройка warehouse_address).
type WarehouseAddress struct {
	Contact string `json:"contact"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	CityCN  string `json:"city_cn"`
}

// GetWarehouseAddress — блоки адреса склада + личный код клиента.
func (s *Service) GetWarehouseAddress(ctx context.Context, clientID int64) (WarehouseAddress, repo.Client, error) {
	raw, err := s.Settings.Get(ctx, "warehouse_address")
	if errors.Is(err, pgx.ErrNoRows) {
		return WarehouseAddress{}, repo.Client{}, domain.ErrNotFound
	}
	if err != nil {
		return WarehouseAddress{}, repo.Client{}, err
	}
	var wa WarehouseAddress
	if err := json.Unmarshal([]byte(raw), &wa); err != nil {
		return WarehouseAddress{}, repo.Client{}, fmt.Errorf("настройка warehouse_address повреждена: %w", err)
	}
	cl, err := s.Clients.GetByID(ctx, clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return WarehouseAddress{}, repo.Client{}, domain.ErrNotFound
	}
	return wa, cl, err
}
