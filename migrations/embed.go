// Package migrations содержит SQL-миграции, встроенные в бинарь (авто-применение при старте api).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
