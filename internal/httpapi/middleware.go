package httpapi

import (
	"context"
	"net/http"
	"strings"

	"cargo/internal/auth"
)

type ctxKey string

const clientIDKey ctxKey = "clientID"

// ClientAuth — middleware Bearer JWT для клиентских роутов.
func ClientAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				writeErr(w, http.StatusUnauthorized, "unauthorized", "нет токена")
				return
			}
			claims, err := auth.ParseToken(secret, token)
			if err != nil || claims.Role != auth.RoleClient {
				writeErr(w, http.StatusUnauthorized, "unauthorized", "недействительный токен")
				return
			}
			ctx := context.WithValue(r.Context(), clientIDKey, claims.ClientID())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// clientIDFrom — id клиента из контекста.
func clientIDFrom(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(clientIDKey).(int64)
	return id, ok
}
