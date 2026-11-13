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

const (
	adminIDKey ctxKey = "adminID"
	adminRole  ctxKey = "adminRole"
)

// AdminAuth — middleware Bearer JWT для админских роутов (operator/admin).
func AdminAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				writeErr(w, http.StatusUnauthorized, "unauthorized", "нет токена")
				return
			}
			claims, err := auth.ParseToken(secret, token)
			if err != nil || (claims.Role != auth.RoleOperator && claims.Role != auth.RoleAdmin) {
				writeErr(w, http.StatusUnauthorized, "unauthorized", "недействительный токен")
				return
			}
			ctx := context.WithValue(r.Context(), adminIDKey, claims.ClientID()) // subject = id оператора
			ctx = context.WithValue(ctx, adminRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdminRole — доступ только для роли admin (тарифы, settings).
func RequireAdminRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if role, _ := r.Context().Value(adminRole).(string); role != auth.RoleAdmin {
			writeErr(w, http.StatusForbidden, "forbidden", "доступно только роли admin")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// adminIDFrom — id оператора из контекста (для audit-событий).
func adminIDFrom(ctx context.Context) int64 {
	id, _ := ctx.Value(adminIDKey).(int64)
	return id
}
