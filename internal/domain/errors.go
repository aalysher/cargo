package domain

import "errors"

// ErrNotFound — объект не найден (HTTP 404).
var ErrNotFound = errors.New("not found")

// ErrUnauthorized — неверные учётные данные (HTTP 401).
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden — недостаточно прав (HTTP 403).
var ErrForbidden = errors.New("forbidden")

// ErrInvalidTransition — недопустимый переход статуса (HTTP 409).
var ErrInvalidTransition = errors.New("invalid status transition")

// ValidationError — некорректный ввод (HTTP 400, Detail — причина для пользователя).
type ValidationError struct{ Detail string }

// Error — сообщение об ошибке валидации.
func (e ValidationError) Error() string { return "invalid input: " + e.Detail }

// ValidationErr — конструктор ошибки валидации.
func ValidationErr(detail string) error { return ValidationError{Detail: detail} }
