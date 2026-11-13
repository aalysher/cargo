package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"

	"cargo/internal/domain"
)

// readJSON декодирует JSON-тело (лимит 1 МБ, неизвестные поля запрещены).
func (h *Handler) readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", "некорректный JSON")
		return false
	}
	return true
}

// writeJSON отправляет JSON-ответ.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr отправляет ошибку в едином формате.
func writeErr(w http.ResponseWriter, status int, code, detail string) {
	writeJSON(w, status, errResp{Error: code, Detail: detail})
}

// fail маппит ошибки в HTTP-ответы.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		writeErr(w, http.StatusNotFound, "not_found", "не найдено")
	default:
		var ve domain.ValidationError
		if errors.As(err, &ve) {
			writeErr(w, http.StatusBadRequest, "validation_error", ve.Detail)
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeErr(w, http.StatusConflict, "invalid_transition", err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			writeErr(w, http.StatusUnauthorized, "unauthorized", "неверный логин или пароль")
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			writeErr(w, http.StatusForbidden, "forbidden", "недостаточно прав")
			return
		}
		slog.ErrorContext(r.Context(), "internal error", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
	}
}
