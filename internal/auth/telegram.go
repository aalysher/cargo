// Package auth: проверка Telegram initData и JWT-сессии.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidInitData — initData отсутствует, с неверной подписью или устарел.
var ErrInvalidInitData = errors.New("invalid telegram init data")

// TelegramUser — объект user из initData.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

// ValidateInitData проверяет подпись initData по официальному алгоритму Telegram:
// data_check_string (все поля, кроме hash, отсортированы) → HMAC-SHA256 с ключом
// HMAC(bot_token, "WebAppData"). maxAge <= 0 — проверка свежести отключена (dev).
func ValidateInitData(initData, botToken string, maxAge time.Duration) (TelegramUser, error) {
	if initData == "" || botToken == "" {
		return TelegramUser{}, ErrInvalidInitData
	}
	vals, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	hash := vals.Get("hash")
	if hash == "" {
		return TelegramUser{}, ErrInvalidInitData
	}

	parts := make([]string, 0, len(vals))
	for k, vs := range vals {
		if k == "hash" || len(vs) == 0 {
			continue
		}
		parts = append(parts, k+"="+vs[0])
	}
	sort.Strings(parts)

	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	expected := hex.EncodeToString(hmacSHA256(secret, []byte(strings.Join(parts, "\n"))))
	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return TelegramUser{}, ErrInvalidInitData
	}

	if maxAge > 0 {
		authDate, err := strconv.ParseInt(vals.Get("auth_date"), 10, 64)
		if err != nil || time.Now().Unix()-authDate > int64(maxAge/time.Second) {
			return TelegramUser{}, ErrInvalidInitData
		}
	}

	var user TelegramUser
	if err := json.Unmarshal([]byte(vals.Get("user")), &user); err != nil || user.ID == 0 {
		return TelegramUser{}, ErrInvalidInitData
	}
	return user, nil
}

func hmacSHA256(key, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	return h.Sum(nil)
}
