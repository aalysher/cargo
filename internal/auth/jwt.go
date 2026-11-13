package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Роли в JWT.
const (
	RoleClient   = "client"
	RoleOperator = "operator"
	RoleAdmin    = "admin"
)

// clientSessionTTL — срок жизни клиентской сессии.
const clientSessionTTL = 30 * 24 * time.Hour

// adminSessionTTL — срок жизни сессии оператора/админа (короткая, безопасность).
const adminSessionTTL = 12 * time.Hour

// Claims — данные сессии (subject = id пользователя, role = роль).
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// NewClientToken выпускает JWT для клиента.
func NewClientToken(secret string, clientID int64) (string, error) {
	claims := Claims{
		Role: RoleClient,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(clientID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(clientSessionTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// NewAdminToken выпускает JWT для оператора/админа (роль задаётся явно).
func NewAdminToken(secret string, userID int64, role string) (string, error) {
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(adminSessionTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken разбирает и валидирует JWT.
func ParseToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ClientID возвращает id клиента из subject.
func (c *Claims) ClientID() int64 {
	id, _ := strconv.ParseInt(c.Subject, 10, 64)
	return id
}
