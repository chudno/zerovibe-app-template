// Package sqlite — реализация портов usecase поверх SQLite.
//
// КЛЮЧЕВОЙ ПАТТЕРН: записи (Create/Delete/Update) идут через db.Write — попадают в
// единую writer-горутину и сериализуются (нет SQLITE_BUSY). Чтения (List/Get) идут
// через db.Read — параллельно. Конвертация domain↔строка БД — в этом слое.
//
// ОБРАЗЕЦ ДЛЯ ГЕНЕРАЦИИ: на каждую сущность — свой репозиторий, реализующий
// порт из usecase. INSERT/UPDATE/DELETE оборачивать в db.Write, SELECT — в db.Read.
package sqlite

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// writer — минимальный интерфейс к платформенному db.DB (Write/Read). Принимаем
// интерфейсом, а не *db.DB, чтобы репозитории было легко тестировать.
type writer interface {
	Write(ctx context.Context, fn func(*sql.DB) error) error
	Read(fn func(*sql.DB) error) error
}

// sqliteTime — формат хранения времени в текстовых колонках (как datetime('now')).
// Единый формат для записи и чтения по всем репозиториям.
const sqliteTime = "2006-01-02 15:04:05"

// parseTime разбирает datetime('now')-строку SQLite в time.Time. При ошибке
// возвращает нулевое время, но ЛОГИРУЕТ — иначе битый формат тихо делал бы
// сессию/токен мгновенно невалидными (нулевое время = «истекло»), и причину было
// бы не найти. Пустая строка (не заданное значение) — норма, не логируем.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(sqliteTime, s)
	if err != nil {
		slog.Warn("sqlite: не удалось разобрать время", "value", s, "error", err)
		return time.Time{}
	}
	return t
}
