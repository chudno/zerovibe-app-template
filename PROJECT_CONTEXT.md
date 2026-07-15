# PROJECT_CONTEXT — код шаблона одним файлом

Автоген (`make context`). Здесь путь и содержимое каждого файла шаблона —
читай ЭТОТ файл, чтобы понять структуру и паттерны, вместо изучения файлов
по отдельности. Пересобирается при изменении шаблона; НЕ правь вручную.

## Файлы (57)

- `Makefile`
- `assets/input.css`
- `cmd/server/main.go`
- `go.mod`
- `internal/adapter/platformfiles/platformfiles.go`
- `internal/adapter/platformmail/platformmail.go`
- `internal/admin/icons.go`
- `internal/admin/query.go`
- `internal/admin/registry.go`
- `internal/admin/resource.go`
- `internal/admin/server.go`
- `internal/admin/templates/form.html`
- `internal/admin/templates/layout.html`
- `internal/admin/templates/list.html`
- `internal/admin/templates/login.html`
- `internal/admin/view.go`
- `internal/adminres/note.go`
- `internal/adminres/user.go`
- `internal/domain/email_verification.go`
- `internal/domain/errors.go`
- `internal/domain/note.go`
- `internal/domain/reset.go`
- `internal/domain/session.go`
- `internal/domain/setting.go`
- `internal/domain/user.go`
- `internal/platform/db/db.go`
- `internal/platform/db/migrate.go`
- `internal/platform/db/migrations/00001_init.sql`
- `internal/platform/db/migrations/00002_settings.sql`
- `internal/platform/db/migrations/00003_auth.sql`
- `internal/platform/db/migrations/00004_email_verification.sql`
- `internal/platform/db/migrations/00005_note_due_date.sql`
- `internal/repository/sqlite/email_verifications.go`
- `internal/repository/sqlite/notes.go`
- `internal/repository/sqlite/notes_admin.go`
- `internal/repository/sqlite/ratelimit.go`
- `internal/repository/sqlite/resets.go`
- `internal/repository/sqlite/sessions.go`
- `internal/repository/sqlite/settings.go`
- `internal/repository/sqlite/sqlite.go`
- `internal/repository/sqlite/users.go`
- `internal/repository/sqlite/users_admin.go`
- `internal/transport/web/templates/forgot.html`
- `internal/transport/web/templates/landing.html`
- `internal/transport/web/templates/layout.html`
- `internal/transport/web/templates/login.html`
- `internal/transport/web/templates/note.html`
- `internal/transport/web/templates/notes.html`
- `internal/transport/web/templates/register.html`
- `internal/transport/web/templates/reset.html`
- `internal/transport/web/templates/settings.html`
- `internal/transport/web/templates/verify.html`
- `internal/transport/web/web.go`
- `internal/usecase/auth.go`
- `internal/usecase/files.go`
- `internal/usecase/notes.go`
- `internal/usecase/settings.go`

---

## `Makefile`

```make
.PHONY: run dev build test vet check css css-watch docker docker-run tidy context clean

# Версия standalone-бинаря tailwindcss-extra (Tailwind CSS + DaisyUI внутри,
# Node НЕ нужен). Пин версии — для воспроизводимости сборки.
TW_VERSION := v2.9.1
TW_BIN := ./bin/tailwindcss-extra
# Ассет под текущую ОС/арх (локально — macOS; в Docker переопределяется).
TW_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/macos/')
TW_ARCH := $(shell uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/')
TW_ASSET := tailwindcss-extra-$(TW_OS)-$(TW_ARCH)
TW_URL := https://github.com/dobicinaitis/tailwind-cli-extra/releases/download/$(TW_VERSION)/$(TW_ASSET)

CSS_IN := assets/input.css
CSS_OUT := internal/transport/web/static/app.css

# Скачать бинарь tailwindcss-extra, если ещё нет.
$(TW_BIN):
	@mkdir -p bin
	@echo "→ качаю $(TW_ASSET) ($(TW_VERSION))"
	@curl -sSL -o $(TW_BIN) $(TW_URL)
	@chmod +x $(TW_BIN)

# Сборка CSS: сканирует html-шаблоны, генерит минимальный CSS (tree-shaking) в
# static/app.css, который вшивается в бинарь через embed. Гоняй после правки
# классов в шаблонах.
css: $(TW_BIN)
	$(TW_BIN) -i $(CSS_IN) -o $(CSS_OUT) --minify

# Watch-режим для разработки (пересобирает CSS при изменении шаблонов).
css-watch: $(TW_BIN)
	$(TW_BIN) -i $(CSS_IN) -o $(CSS_OUT) --watch

# Локальный запуск (БД в текущем каталоге). CSS собираем перед стартом.
run: css
	go run ./cmd/server

# Dev-режим live-reload: ZV_DEV=1 заставляет приложение читать html-шаблоны и статику
# С ДИСКА на каждый запрос (а не из вшитого embed). Правки .html и app.css видны сразу
# по F5, без пересборки бинаря. Параллельно держи `make css-watch` в другом терминале —
# тогда и стили пересобираются на лету. Правки .go всё же требуют перезапуска (Ctrl-C +
# `make dev`) — embed тут ни при чём, просто бинарь надо собрать заново.
dev: css
	ZV_DEV=1 SECURE_COOKIE=false go run ./cmd/server

# Проверка перед публикацией: CSS + сборка + статанализ + тесты.
check: css
	go build ./... && go vet ./... && go test ./...

# Сборка бинаря (CSS собирается первым — попадает в embed).
build: css
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/zerovibe ./cmd/server

# Тесты (unit + e2e, без сети)
test:
	go test ./...

vet:
	go vet ./...

# Сборка Docker-образа
docker:
	docker build -t zerovibe:local .

# Запуск контейнера с volume под данные (порт 8080)
docker-run: docker
	docker run --rm -p 8080:8080 -v zerovibe-data:/data zerovibe:local

tidy:
	go mod tidy

# context — собрать весь код шаблона в PROJECT_CONTEXT.md (путь+содержимое каждого
# файла), чтобы агент читал ОДИН файл на старте вместо десятков Read/Glob. Пере-
# собирать при изменении структуры шаблона.
context:
	@bash scripts/gen-context.sh

clean:
	rm -rf bin zerovibe.db zerovibe.db-wal zerovibe.db-shm
```

## `assets/input.css`

```css
/* Исходник CSS приложения — Tailwind CSS v4 + DaisyUI (собирается в
 * internal/transport/web/static/app.css через `make css`; результат вшивается в
 * бинарь через embed). Собирается standalone-бинарём tailwindcss-extra (Node НЕ
 * нужен, DaisyUI внутри бинаря).
 *
 * source(none) + явный @source: сканируем ТОЛЬКО html-шаблоны приложения —
 * tree-shaking, в итоговый CSS попадают лишь реально использованные классы и
 * компоненты DaisyUI (обычно 5–15 kB вместо всей библиотеки).
 *
 * Темы: garden по умолчанию (всегда, независимо от системной темы ОС). */
@import "tailwindcss" source(none);
@plugin "daisyui" {
	/* Включены ВСЕ встроенные темы daisyUI. Пользователь выбирает тему
	   переключателем в шапке (см. "theme-toggle" в layout.html); выбор
	   сохраняется. По умолчанию (без явного выбора) — garden, ВСЕГДА, независимо
	   от системной темы ОС. Метку --prefersdark намеренно НЕ ставим: иначе на
	   устройстве с тёмной системной темой она перебивала бы дефолт (--default).
	   Чтобы сделать дефолтной другую тему — перенеси метку --default на неё. */
	themes: garden --default, dark, light, cupcake, bumblebee, emerald,
		corporate, synthwave, retro, cyberpunk, valentine, halloween,
		forest, aqua, lofi, pastel, fantasy, wireframe, black, luxury, dracula,
		cmyk, autumn, business, acid, lemonade, night, coffee, winter, dim,
		nord, sunset, caramellatte, abyss, silk;
}
@source "../internal/transport/web/templates";

/* Шрифты приложения — ЛОКАЛЬНЫЕ (в static/fonts, отдаются с нашего домена, БЕЗ
 * CDN: Google Fonts недоступны из РФ, как и любой внешний хост). Оба под OFL.
 *   - Onest — основной текст: современный геометрично-гуманистический гротеск,
 *     отличная кириллица. Веса 400/500/600.
 *   - Unbounded — ТОЛЬКО заголовки: выразительный дисплейный гротеск, чтобы
 *     заголовки выделялись и лендинг не выглядел «системным». Вес 700.
 * Разбиты на cyrillic/latin сабсеты (unicode-range) — браузер грузит нужный,
 * общий вес ~110 kB. font-display: swap — текст виден сразу системным, затем
 * подменяется (без пустого экрана). Файлы вшиваются в бинарь через embed. */
@font-face {
	font-family: "Onest"; font-style: normal; font-weight: 400; font-display: swap;
	src: url("/static/fonts/onest-400.woff2") format("woff2");
	unicode-range: U+0301, U+0400-045F, U+0490-0491, U+04B0-04B1, U+2116;
}
@font-face {
	font-family: "Onest"; font-style: normal; font-weight: 400; font-display: swap;
	src: url("/static/fonts/onest-latin-400.woff2") format("woff2");
	unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+FEFF;
}
@font-face {
	font-family: "Onest"; font-style: normal; font-weight: 500; font-display: swap;
	src: url("/static/fonts/onest-500.woff2") format("woff2");
	unicode-range: U+0301, U+0400-045F, U+0490-0491, U+04B0-04B1, U+2116;
}
@font-face {
	font-family: "Onest"; font-style: normal; font-weight: 500; font-display: swap;
	src: url("/static/fonts/onest-latin-500.woff2") format("woff2");
	unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+FEFF;
}
@font-face {
	font-family: "Onest"; font-style: normal; font-weight: 600; font-display: swap;
	src: url("/static/fonts/onest-600.woff2") format("woff2");
	unicode-range: U+0301, U+0400-045F, U+0490-0491, U+04B0-04B1, U+2116;
}
@font-face {
	font-family: "Onest"; font-style: normal; font-weight: 600; font-display: swap;
	src: url("/static/fonts/onest-latin-600.woff2") format("woff2");
	unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+FEFF;
}
@font-face {
	font-family: "Unbounded"; font-style: normal; font-weight: 700; font-display: swap;
	src: url("/static/fonts/unbounded-700.woff2") format("woff2");
	unicode-range: U+0301, U+0400-045F, U+0490-0491, U+04B0-04B1, U+2116;
}
@font-face {
	font-family: "Unbounded"; font-style: normal; font-weight: 700; font-display: swap;
	src: url("/static/fonts/unbounded-latin-700.woff2") format("woff2");
	unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+FEFF;
}

/* Onest — основной шрифт всего приложения. Переопределяем Tailwind-переменную
 * --font-sans (её использует DaisyUI и утилиты font-sans) → системный стек как
 * фолбэк, пока Onest грузится / если не загрузился. */
@theme {
	--font-sans: "Onest", ui-sans-serif, system-ui, sans-serif;
}

/* Утилита для заголовков: Unbounded. Вешай class="font-display" на h1/крупные
 * заголовки лендинга, чтобы они выделялись. Обычный текст остаётся на Onest. */
@utility font-display {
	font-family: "Unbounded", ui-sans-serif, system-ui, sans-serif;
}

/* Скругления по умолчанию (темы light/dark) — мягче, чем дефолт DaisyUI (кнопки/
 * поля были почти квадратные, --radius-field:.25rem). Крутим три переменные тем:
 *   --radius-field    — кнопки, инпуты, бейджи, вкладки
 *   --radius-box      — карточки, модалки, крупные блоки
 *   --radius-selector — чекбоксы, тогглы, мелкие контролы
 * Только для дефолтных тем — остальные 33 темы сохраняют свой характер. Светлая
 * идёт и в :root (когда data-theme не выставлен, работает prefers-color-scheme). */
:root,
[data-theme="garden"],
[data-theme="light"],
[data-theme="dark"] {
	--radius-field: 0.65rem;
	--radius-box: 1rem;
	--radius-selector: 0.75rem;
}

/* Плавность hx-boost / htmx-подмен: лёгкое затухание заменяемого контента, чтобы
 * SPA-навигация и подмены фрагментов не «моргали». htmx на время свопа вешает
 * .htmx-swapping на уходящий контент, затем .htmx-settling на новый. Не удалять —
 * на эти классы ссылается skill `conventions` (SPA без перезагрузки). */
.htmx-swapping {
	opacity: 0;
	transition: opacity 90ms ease-out;
}
.htmx-settling {
	opacity: 1;
	transition: opacity 90ms ease-in;
}
```

## `cmd/server/main.go`

```go
// Command server — точка входа эталонного приложения zerovibe.
// Composition root: читает конфиг из окружения, собирает слои
// (db → repository → usecase → transport) и поднимает HTTP-сервер с graceful
// shutdown. Это единственное место, где слои «склеиваются».
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/chudno/zerovibe/internal/adapter/platformfiles"
	"github.com/chudno/zerovibe/internal/adapter/platformmail"
	"github.com/chudno/zerovibe/internal/admin"
	"github.com/chudno/zerovibe/internal/adminres"
	"github.com/chudno/zerovibe/internal/platform/db"
	"github.com/chudno/zerovibe/internal/repository/sqlite"
	"github.com/chudno/zerovibe/internal/transport/web"
	"github.com/chudno/zerovibe/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := env("ADDR", ":8080")
	dbPath := env("DB_PATH", "file:zerovibe.db")
	appBaseURL := env("APP_BASE_URL", "http://localhost:8080")
	cookieName := env("SESSION_COOKIE", "zv_session")
	secureCookie := envBool("SECURE_COOKIE", true)
	// ZV_PREVIEW=1 — приложение запущено в live-превью платформы (cross-site iframe):
	// сессионную cookie надо ставить SameSite=None; Secure, иначе вход не удержится.
	previewMode := envBool("ZV_PREVIEW", false)
	// Платформа подставляет адрес API и сервис-ключ для отправки писем при деплое.
	platformURL := os.Getenv("PLATFORM_API_URL")
	platformKey := os.Getenv("PLATFORM_API_KEY")
	// Первый админ создаётся через POST /setup по одноразовому коду SETUP_TOKEN.
	// Код задаёт и передаёт платформа при деплое (env); локально его можно задать
	// в .env. Это единственный путь — отдельного сида из env-кредов нет.
	setupToken := os.Getenv("SETUP_TOKEN")

	// db (платформенный слой: SQLite + очередь записи)
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	ctx := context.Background()
	if err := database.MigrateUp(ctx); err != nil {
		return err
	}

	// repository
	noteRepo := sqlite.NewNoteRepo(database)
	userRepo := sqlite.NewUserRepo(database)
	sessRepo := sqlite.NewSessionRepo(database)
	resetRepo := sqlite.NewResetRepo(database)
	verifyRepo := sqlite.NewEmailVerificationRepo(database)
	rlRepo := sqlite.NewRateLimitRepo(database)
	settingRepo := sqlite.NewSettingRepo(database)

	// adapters
	mailer := platformmail.New(platformURL, platformKey)
	// Файловое хранилище: прод — presigned-URL платформы (S3); локально — каталог
	// /data/uploads (DB_PATH рядом), раздаётся как /uploads/*.
	files := platformfiles.New(platformURL, platformKey, env("UPLOADS_DIR", "uploads"))
	hasher := usecase.NewBcryptHasher()

	// usecase
	settings := usecase.NewSettingsService(settingRepo)
	notes := usecase.NewNoteService(noteRepo)
	auth := usecase.NewAuthService(
		userRepo, sessRepo, resetRepo, verifyRepo, rlRepo,
		hasher, mailer, settings,
		usecase.AuthConfig{
			SessionTTL:      30 * 24 * time.Hour,
			ResetTTL:        time.Hour,
			VerifyTTL:       24 * time.Hour,
			AppBaseURL:      appBaseURL,
			LoginRateLimit:  usecase.RateRule{Limit: 5, Window: 15 * time.Minute},
			ForgotRateLimit: usecase.RateRule{Limit: 3, Window: time.Hour},
			ResendShortRate: usecase.RateRule{Limit: 1, Window: time.Minute},
			ResendHourRate:  usecase.RateRule{Limit: 5, Window: time.Hour},
			SetupToken:      setupToken,
		},
	)

	// Первый администратор создаётся через /setup по коду SETUP_TOKEN: код задаёт и
	// передаёт платформа при деплое (env), приложение его только читает. После деплоя
	// вайбкодер открывает /setup и задаёт свой email/пароль. /setup работает, только
	// пока админа ещё нет — после создания первого закрывается навсегда.
	if needed, err := auth.SetupNeeded(ctx); err != nil {
		return err
	} else if needed {
		log.Print("ПЕРВИЧНАЯ НАСТРОЙКА: админ ещё не создан. Создайте его вызовом POST /setup " +
			"с полями email, password и кодом настройки (SETUP_TOKEN).")
	}

	// transport
	srv, err := web.NewServer(notes, auth, settings, web.Config{
		SecureCookie: secureCookie,
		CookieName:   cookieName,
		PreviewMode:  previewMode,
	})
	if err != nil {
		return err
	}

	// Встроенная админка: реестр сущностей + generic-CRUD. Регистрируем сущности
	// приложения (эталон — Note и User). Добавить новую сущность в админку = одна
	// строка adminres.RegisterX(reg, repo) здесь (см. skill new-feature).
	reg := admin.NewRegistry()
	adminres.RegisterUser(reg, userRepo, hasher)
	adminres.RegisterNote(reg, noteRepo, adminres.UserOptions(userRepo))
	adminSrv, err := admin.NewServer(reg, files)
	if err != nil {
		return err
	}
	srv.SetAdmin(adminSrv)

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// graceful shutdown по SIGINT/SIGTERM
	go func() {
		log.Printf("zerovibe слушает %s (db=%s)", addr, dbPath)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("остановка...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpSrv.Shutdown(shutdownCtx)
}

// env возвращает значение переменной окружения или fallback.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envBool парсит булеву переменную окружения (1/true/yes/on → true); иначе fallback.
func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	switch v {
	case "yes", "on", "YES", "ON":
		return true
	case "no", "off", "NO", "OFF":
		return false
	}
	return fallback
}
```

## `go.mod`

```go
module github.com/chudno/zerovibe

go 1.25.4

require (
	github.com/pressly/goose/v3 v3.27.0
	golang.org/x/crypto v0.48.0
	modernc.org/sqlite v1.51.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	modernc.org/libc v1.72.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)
```

## `internal/adapter/platformfiles/platformfiles.go`

```go
// Package platformfiles — хранение файлов через платформенный файловый API Zerovibe.
// Реализует порт usecase.FileStore. Платформа выдаёт presigned-URL на S3 (как с
// почтой: сами S3-креды приложению не передаются), приложение льёт/читает байты по
// этим URL. Локально (без ключа платформы) — фолбэк в каталог на диске, чтобы
// загрузка файлов работала в dev без настроенного S3.
package platformfiles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Client — файловое хранилище поверх платформенного API.
type Client struct {
	apiURL  string
	apiKey  string
	http    *http.Client
	localDir string // каталог фолбэка, когда платформа не настроена (локальная разработка)
	logf    func(string, ...any)
}

// New собирает клиент. apiURL/apiKey приходят из окружения (PLATFORM_API_URL/KEY) —
// те же, что у почты. localDir — куда складывать файлы в dev-режиме (без ключа).
func New(apiURL, apiKey, localDir string) *Client {
	return &Client{
		apiURL:   strings.TrimRight(apiURL, "/"),
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 30 * time.Second},
		localDir: localDir,
		logf:     log.Printf,
	}
}

// configured сообщает, настроена ли платформа (есть ключ и адрес). Иначе — фолбэк.
func (c *Client) configured() bool { return c.apiKey != "" && c.apiURL != "" }

// Save сохраняет файл и возвращает его КЛЮЧ (приложение хранит ключ в своей записи).
// Прод: просит у платформы presigned PUT-URL и заливает на него байты. Dev (нет
// ключа): пишет в localDir и возвращает локальный ключ "local/<имя>".
func (c *Client) Save(ctx context.Context, fileName string, content io.Reader, size int64) (string, error) {
	if !c.configured() {
		return c.saveLocal(fileName, content)
	}

	// 1) presigned PUT-URL + ключ от платформы.
	up, err := c.requestUpload(ctx, fileName)
	if err != nil {
		return "", err
	}

	// 2) PUT байтов файла на выданный URL (напрямую в S3, минуя платформу).
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, up.UploadURL, content)
	if err != nil {
		return "", fmt.Errorf("platformfiles: put request: %w", err)
	}
	if size > 0 {
		req.ContentLength = size
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("platformfiles: upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("platformfiles: upload status %d", resp.StatusCode)
	}
	return up.Key, nil
}

// URL возвращает ссылку для показа/скачивания файла по его ключу. Прод: presigned
// GET-URL от платформы. Dev: путь "/uploads/<имя>", раздаётся приложением со static.
func (c *Client) URL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	if !c.configured() {
		return c.localURL(key), nil
	}
	body, _ := json.Marshal(map[string]string{"key": key})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/dev/files/get", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("platformfiles: get request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("platformfiles: get url: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("platformfiles: get url status %d", resp.StatusCode)
	}
	var out struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("platformfiles: decode get url: %w", err)
	}
	return out.URL, nil
}

// uploadResp — ответ платформы на запрос заливки.
type uploadResp struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
}

// requestUpload просит у платформы presigned PUT-URL и ключ для имени файла.
func (c *Client) requestUpload(ctx context.Context, fileName string) (uploadResp, error) {
	body, _ := json.Marshal(map[string]string{"file_name": fileName})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/dev/files", bytes.NewReader(body))
	if err != nil {
		return uploadResp{}, fmt.Errorf("platformfiles: upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return uploadResp{}, fmt.Errorf("platformfiles: request upload url: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return uploadResp{}, fmt.Errorf("platformfiles: request upload url status %d", resp.StatusCode)
	}
	var out uploadResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return uploadResp{}, fmt.Errorf("platformfiles: decode upload url: %w", err)
	}
	return out, nil
}

// --- Локальный фолбэк (dev без платформы) ---

// saveLocal пишет файл в localDir под безопасным базовым именем и возвращает ключ
// "local/<имя>". Раздаётся приложением как /uploads/<имя> (см. localURL).
func (c *Client) saveLocal(fileName string, content io.Reader) (string, error) {
	name := safeBase(fileName)
	if err := os.MkdirAll(c.localDir, 0o755); err != nil {
		return "", fmt.Errorf("platformfiles: mkdir local: %w", err)
	}
	dst := filepath.Join(c.localDir, name)
	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("platformfiles: create local: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("platformfiles: write local: %w", err)
	}
	c.logf("[platformfiles] платформа не настроена — файл сохранён локально: %s", dst)
	return "local/" + name, nil
}

// localURL переводит локальный ключ в путь, который приложение раздаёт со static.
func (c *Client) localURL(key string) string {
	return "/uploads/" + strings.TrimPrefix(key, "local/")
}

// safeBase возвращает безопасное базовое имя файла (без каталогов и "../").
func safeBase(name string) string {
	name = strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	name = path.Base(name)
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	return name
}
```

## `internal/adapter/platformmail/platformmail.go`

```go
// Package platformmail — отправка писем через платформенный email-API Zerovibe.
// Реализует порт usecase.Mailer. Платформа сама подставляет адрес API и сервис-ключ
// в окружение контейнера при деплое; локально/в тестах ключа может не быть — тогда
// письмо логируется (ссылка видна в консоли), а поток не падает.
package platformmail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chudno/zerovibe/internal/usecase"
)

// Client — отправитель писем поверх платформенного API.
type Client struct {
	apiURL string
	apiKey string
	http   *http.Client
	logf   func(string, ...any)
}

// New собирает клиент. apiURL/apiKey приходят из окружения (PLATFORM_API_URL/KEY).
// Пустой apiKey включает режим лога (см. Send).
func New(apiURL, apiKey string) *Client {
	return &Client{
		apiURL: strings.TrimRight(apiURL, "/"),
		apiKey: apiKey,
		http:   &http.Client{Timeout: 10 * time.Second},
		logf:   log.Printf,
	}
}

// Send отправляет письмо. Без ключа (локальная разработка) — логирует и возвращает
// nil, чтобы восстановление пароля работало в dev без настроенной почты.
func (c *Client) Send(ctx context.Context, m usecase.Email) error {
	if c.apiKey == "" || c.apiURL == "" {
		// Грейсфул-фолбэк ТОЛЬКО для локальной разработки (ключа нет): печатаем письмо
		// со ссылкой в консоль, чтобы можно было пройти флоу без настроенной почты.
		// В проде ключ всегда прокинут платформой → эта ветка не выполняется, тело
		// письма (со ссылкой-токеном) в логи не попадает.
		c.logf("[platformmail] почта не настроена — письмо не отправлено. Кому: %s. Тема: %s. Текст:\n%s",
			m.To, m.Subject, m.Text)
		return nil
	}

	body, err := json.Marshal(map[string]string{
		"to":      m.To,
		"subject": m.Subject,
		"text":    m.Text,
		"html":    m.HTML,
	})
	if err != nil {
		return fmt.Errorf("platformmail: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/dev/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("platformmail: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("platformmail: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("platformmail: статус %d от email-API", resp.StatusCode)
	}
	return nil
}
```

## `internal/admin/icons.go`

```go
package admin

import "html/template"

// Иконки админки — Lucide (https://lucide.dev), встроены ЛОКАЛЬНО как inline-SVG
// (без CDN/JS-библиотеки: важно для доступности из РФ и минимализма стэка). Здесь
// только внутренности <svg> (paths) — обёртку <svg> со штрихом рисует функция icon.
//
// Толщина штриха 2.25 («пожирнее» — это служебная админка, читаемость важнее изящества).
// Добавить иконку = взять path из lucide.dev и положить сюда по ключу.
var lucidePaths = map[string]template.HTML{
	// действия
	"pencil":  `<path d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z"/><path d="m15 5 4 4"/>`,
	"trash":   `<path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/>`,
	"x":       `<path d="M18 6 6 18"/><path d="m6 6 12 12"/>`,
	"plus":    `<path d="M5 12h14"/><path d="M12 5v14"/>`,
	"search":  `<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>`,
	"download": `<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/>`,
	"chevron-left":  `<path d="m15 18-6-6 6-6"/>`,
	"chevron-right": `<path d="m9 18 6-6-6-6"/>`,
	"arrow-up-down": `<path d="m21 16-4 4-4-4"/><path d="M17 20V4"/><path d="m3 8 4-4 4 4"/><path d="M7 4v16"/>`,
	"arrow-up":      `<path d="m5 12 7-7 7 7"/><path d="M12 19V5"/>`,
	"arrow-down":    `<path d="M12 5v14"/><path d="m19 12-7 7-7-7"/>`,
	"log-out":  `<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/>`,
	"check":    `<path d="M20 6 9 17l-5-5"/>`,
	"minus":    `<path d="M5 12h14"/>`,
	"upload":   `<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/>`,
	"file":     `<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/>`,
	"shield":   `<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/>`,

	// иконки разделов (сайдбар) — дефолты для сущностей
	"layout-dashboard": `<rect width="7" height="9" x="3" y="3" rx="1"/><rect width="7" height="5" x="14" y="3" rx="1"/><rect width="7" height="9" x="14" y="12" rx="1"/><rect width="7" height="5" x="3" y="16" rx="1"/>`,
	"users":            `<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>`,
	"file-text":        `<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/>`,
	"list":             `<path d="M3 12h.01"/><path d="M3 18h.01"/><path d="M3 6h.01"/><path d="M8 12h13"/><path d="M8 18h13"/><path d="M8 6h13"/>`,
	"folder":           `<path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/>`,
	"package":          `<path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/>`,
	"tag":              `<path d="M12.586 2.586A2 2 0 0 0 11.172 2H4a2 2 0 0 0-2 2v7.172a2 2 0 0 0 .586 1.414l8.704 8.704a2.426 2.426 0 0 0 3.42 0l6.58-6.58a2.426 2.426 0 0 0 0-3.42z"/><circle cx="7.5" cy="7.5" r=".5" fill="currentColor"/>`,
	"shopping-cart":    `<circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/>`,
	"calendar":         `<path d="M8 2v4"/><path d="M16 2v4"/><rect width="18" height="18" x="3" y="4" rx="2"/><path d="M3 10h18"/>`,
	"settings":         `<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>`,
}

// icon рендерит inline-SVG-иконку по имени из набора Lucide. size — сторона в px.
// Неизвестное имя → дефолт (точка), чтобы не падать. Цвет наследуется (currentColor).
func icon(name string, size int) template.HTML {
	if size <= 0 {
		size = 16
	}
	body, ok := lucidePaths[name]
	if !ok {
		body = `<circle cx="12" cy="12" r="3"/>`
	}
	return template.HTML(`<svg xmlns="http://www.w3.org/2000/svg" width="` +
		itoa(size) + `" height="` + itoa(size) +
		`" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.25" stroke-linecap="round" stroke-linejoin="round" style="display:inline-block;vertical-align:middle">` +
		string(body) + `</svg>`)
}

// itoa — маленький helper (без strconv-импорта ради одной функции в шаблонном слое).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
```

## `internal/admin/query.go`

```go
package admin

import (
	"strconv"
	"strings"
)

// compareCells сравнивает два значения ячеек для сортировки. Числовые поля — как
// числа (иначе "10" < "9" лексикографически), остальные — регистронезависимо строкой.
func compareCells(a, b string, t FieldType) bool {
	if t == FieldNumber {
		af, aerr := strconv.ParseFloat(strings.TrimSpace(a), 64)
		bf, berr := strconv.ParseFloat(strings.TrimSpace(b), 64)
		if aerr == nil && berr == nil {
			return af < bf
		}
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

// ListQuery — параметры запроса списка (из query-string): поиск, фильтры, сортировка,
// страница. Generic-слой применяет их к записям ПОСЛЕ List() (фильтрация в памяти —
// для SQLite-приложений объёмы небольшие; точечная оптимизация — потом в репозитории).
type ListQuery struct {
	Search   string            // подстрока поиска по текстовым полям
	Filters  map[string]string // field name -> выбранное значение (select/badge)
	SortBy   string            // имя поля сортировки
	SortDir  string            // "asc" | "desc"
	Page     int               // 1-based
	PageSize int
}

// pageWindow — результат пагинации: срез записей текущей страницы и метаданные.
type pageWindow struct {
	Records  []Record
	Page     int
	Pages    int
	Total    int
	PageSize int
}

// apply фильтрует, сортирует и нарезает записи на страницу по запросу q и полям h.
func (q ListQuery) apply(h ResourceHandle, records []Record) pageWindow {
	records = filterRecords(h, records, q)
	if q.SortBy != "" {
		if f, ok := fieldByName(h, q.SortBy); ok && f.Sortable {
			sortRecords(records, f, q.SortDir)
		}
	}

	total := len(records)
	size := q.PageSize
	if size <= 0 {
		size = 20
	}
	pages := (total + size - 1) / size
	if pages == 0 {
		pages = 1
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	if page > pages {
		page = pages
	}
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	return pageWindow{
		Records:  records[start:end],
		Page:     page,
		Pages:    pages,
		Total:    total,
		PageSize: size,
	}
}

// filterRecords применяет поиск (по InList-текстовым полям) и точные фильтры.
func filterRecords(h ResourceHandle, records []Record, q ListQuery) []Record {
	search := strings.ToLower(strings.TrimSpace(q.Search))
	if search == "" && len(q.Filters) == 0 {
		return records
	}
	out := make([]Record, 0, len(records))
	for _, rec := range records {
		if search != "" && !recordMatchesSearch(h, rec, search) {
			continue
		}
		if !recordMatchesFilters(rec, q.Filters) {
			continue
		}
		out = append(out, rec)
	}
	return out
}

// recordMatchesSearch — true, если поиск-подстрока встречается в любой показываемой
// в списке текстовой ячейке (по Display — то, что видит пользователь).
func recordMatchesSearch(h ResourceHandle, rec Record, search string) bool {
	for _, f := range h.Fields() {
		if !f.InList {
			continue
		}
		c := rec.Cells[f.Name]
		if strings.Contains(strings.ToLower(c.Display), search) || strings.Contains(strings.ToLower(c.Value), search) {
			return true
		}
	}
	return false
}

// recordMatchesFilters — true, если запись удовлетворяет всем точным фильтрам.
func recordMatchesFilters(rec Record, filters map[string]string) bool {
	for name, want := range filters {
		if want == "" {
			continue
		}
		if rec.Cells[name].Value != want {
			return false
		}
	}
	return true
}
```

## `internal/admin/registry.go`

```go
package admin

import (
	"context"
	"sort"
)

// ResourceHandle — дескриптор сущности БЕЗ параметра типа. Generic-хендлеры работают
// только через него: так разнотипные Resource[Note]/Resource[User] лежат в одном
// реестре. Адаптер handle[T] стирает тип T, оборачивая Resource[T].
type ResourceHandle interface {
	Name() string
	Title() string
	Icon() string
	Fields() []Field

	List(ctx context.Context) ([]Record, error)
	Get(ctx context.Context, id string) (FormValues, error)
	Create(ctx context.Context, v FormValues, files map[string]FileInput) error
	Update(ctx context.Context, id string, v FormValues, files map[string]FileInput) error
	Delete(ctx context.Context, id string) error
}

// handle[T] адаптирует типизированный Resource[T] к нетипизированному ResourceHandle.
type handle[T any] struct{ r Resource[T] }

func (h handle[T]) Name() string    { return h.r.Name }
func (h handle[T]) Title() string   { return h.r.Title }
func (h handle[T]) Fields() []Field { return h.r.Fields }

// Icon возвращает имя иконки раздела, по умолчанию "list".
func (h handle[T]) Icon() string {
	if h.r.Icon == "" {
		return "list"
	}
	return h.r.Icon
}

func (h handle[T]) List(ctx context.Context) ([]Record, error) {
	items, err := h.r.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(items))
	for _, it := range items {
		out = append(out, h.r.Row(it))
	}
	return out, nil
}

func (h handle[T]) Get(ctx context.Context, id string) (FormValues, error) {
	return h.r.Get(ctx, id)
}
func (h handle[T]) Create(ctx context.Context, v FormValues, files map[string]FileInput) error {
	return h.r.Create(ctx, v, files)
}
func (h handle[T]) Update(ctx context.Context, id string, v FormValues, files map[string]FileInput) error {
	return h.r.Update(ctx, id, v, files)
}
func (h handle[T]) Delete(ctx context.Context, id string) error {
	return h.r.Delete(ctx, id)
}

// Registry — упорядоченный набор зарегистрированных сущностей. Не глобальный: его
// создаёт composition root (main.go) и передаёт в админ-сервер. Это явная зависимость,
// тестируемая и без скрытого состояния пакета.
type Registry struct {
	order   []string                  // имена в порядке регистрации (порядок в меню)
	byName  map[string]ResourceHandle // быстрый доступ по имени раздела
}

// NewRegistry создаёт пустой реестр.
func NewRegistry() *Registry {
	return &Registry{byName: map[string]ResourceHandle{}}
}

// Register добавляет сущность в реестр. Дженерик-функция (не метод: методы Go не
// могут иметь своих type-параметров). Вызывается в main.go на каждую сущность.
func Register[T any](reg *Registry, r Resource[T]) {
	if _, ok := reg.byName[r.Name]; !ok {
		reg.order = append(reg.order, r.Name)
	}
	reg.byName[r.Name] = handle[T]{r: r}
}

// Get возвращает дескриптор по имени раздела (и флаг существования).
func (reg *Registry) Get(name string) (ResourceHandle, bool) {
	h, ok := reg.byName[name]
	return h, ok
}

// All возвращает дескрипторы в порядке регистрации (для меню сайдбара).
func (reg *Registry) All() []ResourceHandle {
	out := make([]ResourceHandle, 0, len(reg.order))
	for _, name := range reg.order {
		out = append(out, reg.byName[name])
	}
	return out
}

// Empty сообщает, что в реестре нет ни одной сущности (тогда админка скрыта).
func (reg *Registry) Empty() bool { return len(reg.order) == 0 }

// fieldByName ищет поле по имени в дескрипторе (помощник для рендера/сортировки).
func fieldByName(h ResourceHandle, name string) (Field, bool) {
	for _, f := range h.Fields() {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// sortRecords сортирует записи по значению ячейки поля sortField. dir: "asc"/"desc".
// Числовые поля сравниваются как числа, остальные — лексикографически.
func sortRecords(records []Record, f Field, dir string) {
	sort.SliceStable(records, func(i, j int) bool {
		a := records[i].Cells[f.Name].Value
		b := records[j].Cells[f.Name].Value
		less := compareCells(a, b, f.Type)
		if dir == "desc" {
			return !less
		}
		return less
	})
}
```

## `internal/admin/resource.go`

```go
// Package admin — встроенная админ-панель приложения («бэкофис из коробки»).
//
// Идея: ОДИН generic-слой (хендлеры, рендеринг, CSV-экспорт, загрузка файлов)
// обслуживает ЛЮБУЮ сущность. Сущность описывается ДЕСКРИПТОРОМ (Resource) —
// явным набором полей и функций-аксессоров, без рефлексии (в духе zerovibe:
// простой явный код). Это «Django admin без магии reflect».
//
// Чтобы сущность появилась в админке, достаточно зарегистрировать её дескриптор:
//
//	admin.Register(admin.Resource[domain.Note]{
//	    Name: "notes", Title: "Заметки",
//	    Fields: []admin.Field{ ... },
//	    List:   func(ctx) ([]domain.Note, error) { ... },
//	    ...
//	})
//
// Generic-слой не знает о конкретных типах — он работает через RecordSet, который
// дескриптор отдаёт. Типобезопасность обеспечивают замыкания внутри Resource.
package admin

import (
	"context"
	"io"
)

// FieldType — тип поля в форме и в списке. Определяет, как поле редактируется
// (контрол формы) и как отображается в таблице.
type FieldType string

const (
	FieldText     FieldType = "text"     // короткая строка
	FieldTextarea FieldType = "textarea" // многострочный текст
	FieldNumber   FieldType = "number"   // целое/дробное число
	FieldDate     FieldType = "date"     // дата (YYYY-MM-DD)
	FieldBool     FieldType = "bool"     // да/нет (switch в форме, галочка в списке)
	FieldSelect   FieldType = "select"   // выбор из фиксированного списка Options
	FieldRelation FieldType = "relation" // связь с другой сущностью (combobox по имени)
	FieldFile     FieldType = "file"     // загрузка файла (ключ хранится в записи)
	FieldBadge    FieldType = "badge"    // статус-пилюля в списке (значение из Options)
)

// Option — вариант для select/relation/badge: машинное значение и человекочитаемая
// подпись. Для FieldBadge поле Tone задаёт цвет пилюли (см. шаблон).
type Option struct {
	Value string
	Label string
	Tone  string // "", "success", "warning", "destructive", "info" — только для badge
}

// Field — описание одного поля сущности: как его звать, как редактировать и
// показывать. Это чистые ДАННЫЕ (без логики доступа к самой сущности — та в Resource).
type Field struct {
	Name     string    // машинное имя (ключ в форме, заголовок CSV-колонки)
	Label    string    // человекочитаемая подпись
	Type     FieldType // тип контрола/отображения
	Required bool      // обязательное при создании/редактировании
	InList   bool      // показывать колонкой в списке
	Sortable bool      // заголовок колонки кликабелен для сортировки
	Help     string    // подпись-описание под полем в форме

	// Options — варианты для select/badge. Для relation варианты подгружаются
	// динамически (RelationOptions), здесь могут быть пусты.
	Options []Option

	// RelationOptions — для FieldRelation: подгрузка вариантов связи (напр. список
	// пользователей для поля «Владелец»). nil для прочих типов.
	RelationOptions func(ctx context.Context) ([]Option, error)
}

// Cell — отрендеренное значение поля для строки списка: машинное значение (для
// сортировки/CSV) и то, как его показать (Display) с учётом типа (badge tone и т.п.).
type Cell struct {
	Value   string // сырое значение (CSV, сортировка)
	Display string // как показать в таблице (для relation — имя, для file — имя файла)
	Tone    string // для badge — цвет пилюли
	IsFile  bool   // file-ячейка → показать миниатюру/иконку
}

// Record — одна запись сущности в обобщённом виде: идентификатор и ячейки по полям.
type Record struct {
	ID    string
	Cells map[string]Cell // ключ — Field.Name
}

// FormValues — значения формы при создании/редактировании (ключ — Field.Name).
// Для file-полей значение — это io.Reader загруженного файла (см. FileInput).
type FormValues map[string]string

// FileInput — загруженный через форму файл (имя + содержимое + размер).
type FileInput struct {
	FileName string
	Content  io.Reader
	Size     int64
}

// Resource — дескриптор сущности для админки. Параметризован типом сущности T,
// но generic-слой работает с ним через интерфейс ResourceHandle (стирание типа).
//
// Функции-аксессоры (List/Get/Create/Update/Delete/Row) — единственное место, где
// дескриптор знает о конкретном типе T. Их пишет автор сущности (или скилл
// new-feature по эталону). Всё остальное (рендер, формы, CSV, файлы) — общее.
type Resource[T any] struct {
	Name   string  // машинное имя раздела (URL: /admin/<name>), напр. "notes"
	Title  string  // заголовок раздела в меню/шапке, напр. "Заметки"
	Icon   string  // имя Lucide-иконки раздела для сайдбара (пусто → дефолт "list")
	Fields []Field // поля сущности

	// Row превращает сущность в строку списка (ячейки по полям).
	Row func(t T) Record
	// List возвращает все записи (generic-слой сам отфильтрует/отсортирует/постранично).
	List func(ctx context.Context) ([]T, error)
	// Get возвращает одну запись по id (для формы редактирования). Заполняет FormValues.
	Get func(ctx context.Context, id string) (FormValues, error)
	// Create создаёт запись из значений формы и файлов. Возвращает ошибку валидации.
	Create func(ctx context.Context, v FormValues, files map[string]FileInput) error
	// Update обновляет запись id значениями формы и файлами.
	Update func(ctx context.Context, id string, v FormValues, files map[string]FileInput) error
	// Delete удаляет запись по id.
	Delete func(ctx context.Context, id string) error
}
```

## `internal/admin/server.go`

```go
package admin

import (
	"embed"
	"encoding/csv"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/chudno/zerovibe/internal/domain"
	"github.com/chudno/zerovibe/internal/usecase"
)

//go:embed templates/*.html
var templatesFS embed.FS

// maxUploadBytes — потолок размера одной формы с файлами (16 МиБ). Защита от
// переполнения памяти на парсинге multipart.
const maxUploadBytes = 16 << 20

// Server — generic админ-сервер. Держит реестр сущностей, файловое хранилище и свои
// шаблоны. Монтируется веб-сервером приложения под префиксом /admin (см. Routes).
type Server struct {
	reg   *Registry
	files usecase.FileStore
	tmpl  *template.Template
}

// NewServer собирает админ-сервер. files может быть nil (тогда file-поля не работают,
// но остальная админка функционирует). Возвращает nil-безопасную ошибку парсинга шаблонов.
func NewServer(reg *Registry, files usecase.FileStore) (*Server, error) {
	tmpl, err := template.New("admin").Funcs(adminFuncs).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{reg: reg, files: files, tmpl: tmpl}, nil
}

// loginView — данные формы входа в админку.
type loginView struct {
	Email string
	Err   string
}

// RenderLogin рисует отдельную страницу входа в админку (свой дизайн). Вызывается из
// web-сервера (у него auth/сессии); сам admin аутентификацию не делает, только вид.
func (s *Server) RenderLogin(w http.ResponseWriter, email, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "login", loginView{Email: email, Err: errMsg}); err != nil {
		http.Error(w, "внутренняя ошибка админки", http.StatusInternalServerError)
	}
}

// HasResources сообщает, есть ли в админке хоть одна сущность (для решения, монтировать
// ли вход в админку в web-сервере).
func (s *Server) HasResources() bool { return !s.reg.Empty() }

// Mount регистрирует админ-маршруты на переданном mux под префиксом /admin.
// guard оборачивает каждый хендлер проверкой роли (передаётся из web-сервера, чтобы
// admin не дублировал логику аутентификации — она в одном месте приложения).
func (s *Server) Mount(mux *http.ServeMux, guard func(http.HandlerFunc) http.HandlerFunc) {
	if s.reg.Empty() {
		return // нет сущностей — админка не монтируется
	}
	mux.HandleFunc("GET /admin", guard(s.handleHome))
	mux.HandleFunc("GET /admin/{resource}", guard(s.handleList))
	mux.HandleFunc("GET /admin/{resource}/export", guard(s.handleExport))
	mux.HandleFunc("GET /admin/{resource}/new", guard(s.handleNew))
	mux.HandleFunc("POST /admin/{resource}", guard(s.handleCreate))
	mux.HandleFunc("GET /admin/{resource}/{id}/edit", guard(s.handleEdit))
	mux.HandleFunc("PUT /admin/{resource}/{id}", guard(s.handleUpdate))
	mux.HandleFunc("DELETE /admin/{resource}/{id}", guard(s.handleDelete))
}

// resourceParam достаёт дескриптор из пути или пишет 404.
func (s *Server) resourceParam(w http.ResponseWriter, r *http.Request) (ResourceHandle, bool) {
	h, ok := s.reg.Get(r.PathValue("resource"))
	if !ok {
		http.NotFound(w, r)
		return nil, false
	}
	return h, true
}

// handleHome редиректит на первый раздел (у админки нет отдельной «главной»).
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	all := s.reg.All()
	if len(all) == 0 {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/admin/"+all[0].Name(), http.StatusSeeOther)
}

// handleList рендерит экран списка: поиск/фильтры/сортировка/пагинация.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	records, err := h.List(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	q := parseListQuery(r)
	win := q.apply(h, records)
	view := s.listView(r.Context(), h, win, q)

	// Что вернуть, зависит от того, КУДА htmx подменяет (заголовок HX-Target):
	//  - admin-content → переход по разделу из сайдбара: контент раздела (топбар+список);
	//  - listbody      → поиск/фильтр/сортировка/пагинация: только секция списка;
	//  - нет HX        → прямой заход по URL: полная страница.
	// Всё без перезагрузки страницы (кроме первого прямого захода).
	if isHX(r) {
		switch r.Header.Get("HX-Target") {
		case "admin-content":
			s.render(w, "content", view)
		default:
			s.render(w, "listbody", view)
		}
		return
	}
	s.render(w, "list", view)
}

// handleExport отдаёт CSV всех (отфильтрованных) записей раздела.
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	records, err := h.List(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	q := parseListQuery(r)
	records = filterRecords(h, records, q)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+h.Name()+".csv")
	cw := csv.NewWriter(w)
	defer cw.Flush()

	var header []string
	var cols []Field
	for _, f := range h.Fields() {
		if f.InList {
			header = append(header, f.Label)
			cols = append(cols, f)
		}
	}
	_ = cw.Write(header)
	for _, rec := range records {
		row := make([]string, 0, len(cols))
		for _, f := range cols {
			c := rec.Cells[f.Name]
			val := c.Display
			if val == "" {
				val = c.Value
			}
			row = append(row, val)
		}
		_ = cw.Write(row)
	}
}

// handleNew отдаёт форму создания как модалку (htmx подгружает её в контейнер без
// перезагрузки). Прямой заход по URL → форма на полной странице (fallback).
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	s.renderForm(w, r, s.formView(r.Context(), h, "", FormValues{}, ""))
}

// handleCreate обрабатывает отправку формы создания. Успех → закрыть модалку и обновить
// список (через HX-Trigger), без перезагрузки. Ошибка → форма с подписью ошибки обратно.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	values, files, err := s.parseForm(r, h)
	if err == nil {
		err = h.Create(r.Context(), values, files)
	}
	if err != nil {
		s.renderForm(w, r, s.formView(r.Context(), h, "", values, errText(err)))
		return
	}
	s.afterMutation(w, r, h)
}

// handleEdit отдаёт форму редактирования (модалка) с текущими значениями записи.
func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	values, err := h.Get(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.renderForm(w, r, s.formView(r.Context(), h, id, values, ""))
}

// handleUpdate обрабатывает отправку формы редактирования (аналогично create).
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	values, files, err := s.parseForm(r, h)
	if err == nil {
		err = h.Update(r.Context(), id, values, files)
	}
	if err != nil {
		s.renderForm(w, r, s.formView(r.Context(), h, id, values, errText(err)))
		return
	}
	s.afterMutation(w, r, h)
}

// handleDelete удаляет запись. HTMX → обновлённый список без перезагрузки; при ошибке
// валидации (напр. последний админ) — 422 с текстом, который htmx покажет.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	h, ok := s.resourceParam(w, r)
	if !ok {
		return
	}
	if err := h.Delete(r.Context(), r.PathValue("id")); err != nil {
		s.fail(w, err)
		return
	}
	s.afterMutation(w, r, h)
}

// afterMutation отдаёт результат успешной мутации БЕЗ перезагрузки страницы.
//
// Удаление (delete) таргетит #listbody напрямую → возвращаем свежую секцию списка.
// Create/Update таргетят #modal (туда же кладётся форма с ошибкой при провале) → при
// успехе очищаем модалку (пустое тело) и сигналим refreshList, чтобы клиент отдельно
// перезагрузил секцию списка. Для не-htmx (прямой POST) — редирект (fallback).
func (s *Server) afterMutation(w http.ResponseWriter, r *http.Request, h ResourceHandle) {
	if !isHX(r) {
		http.Redirect(w, r, "/admin/"+h.Name(), http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodDelete {
		records, err := h.List(r.Context())
		if err != nil {
			s.fail(w, err)
			return
		}
		q := parseListQuery(r)
		win := q.apply(h, records)
		s.render(w, "listbody", s.listView(r.Context(), h, win, q))
		return
	}
	// create/update: закрыть модалку (пустой #modal) и попросить клиента обновить список.
	w.Header().Set("HX-Trigger", "refreshList")
	w.WriteHeader(http.StatusOK)
}

// renderForm отдаёт форму как модалку (htmx). Прямой заход по URL формы (без htmx) —
// редирект на список: форма открывается только поверх списка, отдельной страницы-формы
// нет (так интерфейс всегда остаётся «как SPA», без самостоятельных form-страниц).
func (s *Server) renderForm(w http.ResponseWriter, r *http.Request, v formView) {
	if !isHX(r) {
		http.Redirect(w, r, "/admin/"+v.Resource, http.StatusSeeOther)
		return
	}
	s.render(w, "formdialog", v)
}

// isHX сообщает, что запрос пришёл от htmx (заголовок HX-Request: true).
func isHX(r *http.Request) bool { return r.Header.Get("HX-Request") == "true" }

// parseForm разбирает форму (multipart, если есть файлы): текстовые значения по полям
// + загруженные файлы. Файлы сохраняются в FileStore сразу, в values кладётся ключ.
func (s *Server) parseForm(r *http.Request, h ResourceHandle) (FormValues, map[string]FileInput, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return FormValues{}, nil, err
	}
	values := FormValues{}
	files := map[string]FileInput{}
	for _, f := range h.Fields() {
		if f.Type == FieldFile {
			if s.files == nil {
				continue
			}
			fh, header, ferr := r.FormFile(f.Name)
			if ferr != nil || header == nil {
				// файл не приложили — оставляем прежний (ключ передаётся скрытым полем)
				values[f.Name] = r.FormValue(f.Name + "_key")
				continue
			}
			key, serr := s.files.Save(r.Context(), header.Filename, fh, header.Size)
			_ = fh.Close()
			if serr != nil {
				return values, files, serr
			}
			values[f.Name] = key
			files[f.Name] = FileInput{FileName: header.Filename, Size: header.Size}
			continue
		}
		if f.Type == FieldBool {
			// чекбокс/switch: присутствие → "true"
			if r.FormValue(f.Name) != "" {
				values[f.Name] = "true"
			} else {
				values[f.Name] = "false"
			}
			continue
		}
		values[f.Name] = strings.TrimSpace(r.FormValue(f.Name))
	}
	return values, files, nil
}

// parseListQuery собирает ListQuery из query-string запроса.
func parseListQuery(r *http.Request) ListQuery {
	q := ListQuery{
		Search:   r.URL.Query().Get("q"),
		Filters:  map[string]string{},
		SortBy:   r.URL.Query().Get("sort"),
		SortDir:  r.URL.Query().Get("dir"),
		PageSize: 20,
	}
	if q.SortDir != "desc" {
		q.SortDir = "asc"
	}
	q.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if q.Page < 1 {
		q.Page = 1
	}
	for k, v := range r.URL.Query() {
		if strings.HasPrefix(k, "f_") && len(v) > 0 {
			q.Filters[strings.TrimPrefix(k, "f_")] = v[0]
		}
	}
	return q
}

// render выполняет именованный шаблон с данными.
func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "внутренняя ошибка админки", http.StatusInternalServerError)
	}
}

// fail мапит ошибку в HTTP-код: доменная валидация/защита → 422, не найдено → 404,
// прочее → 500. Чтобы «нельзя удалить последнего админа» не выглядело сбоем сервера.
func (s *Server) fail(w http.ResponseWriter, err error) {
	var validation domain.ErrValidation
	var notFound domain.ErrNotFound
	switch {
	case errors.As(err, &validation):
		http.Error(w, errText(err), http.StatusUnprocessableEntity)
	case errors.As(err, &notFound):
		http.Error(w, errText(err), http.StatusNotFound)
	default:
		http.Error(w, errText(err), http.StatusInternalServerError)
	}
}

// errText извлекает человекочитаемый текст ошибки (для подписи в форме).
func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
```

## `internal/admin/templates/form.html`

```html
{{/* formdialog — форма в МОДАЛКЕ (htmx). Отправка через hx-post/hx-put: при успехе
     сервер отдаёт обновлённый список в #listbody и шлёт HX-Trigger closeModal (модалка
     закрывается). При ошибке валидации сервер возвращает эту же модалку с подписью. */}}
{{define "formdialog"}}
<div class="modal-overlay" data-close>
  <div class="modal">
    <div class="modal__head">
      <h2>{{if .ID}}Редактирование{{else}}Новая запись{{end}} · {{.Title}}</h2>
      <button class="btn btn--icon" type="button" data-close title="Закрыть">{{icon "x" 16}}</button>
    </div>
    <div class="modal__body">
      <form hx-{{if .ID}}put{{else}}post{{end}}="/admin/{{.Resource}}{{if .ID}}/{{.ID}}{{end}}"
            hx-target="#modal" hx-swap="innerHTML" enctype="multipart/form-data">
        {{if .Err}}<div class="field__error" style="margin-bottom:16px">{{.Err}}</div>{{end}}
        {{template "formfields" .}}
        <div class="formactions">
          <button class="btn btn--ghost" type="button" data-close>Отмена</button>
          <button class="btn btn--default" type="submit">Сохранить</button>
        </div>
      </form>
    </div>
  </div>
</div>
{{end}}


{{/* formfields — сами поля формы. */}}
{{define "formfields"}}
  {{range .Fields}}
    <div class="field{{if .Required}} field--required{{end}}">
      {{if eq (printf "%s" .Type) "bool"}}
        <label class="switchrow">
          <span class="switchrow__text">
            <span class="switchrow__label">{{.Label}}</span>
            {{if .Help}}<span class="switchrow__help">{{.Help}}</span>{{end}}
          </span>
          <input type="checkbox" name="{{.Name}}" value="true"{{if eq .Value "true"}} checked{{end}}>
          <span class="switch__track"><span class="switch__thumb"></span></span>
        </label>

      {{else if eq (printf "%s" .Type) "textarea"}}
        <label for="f-{{.Name}}">{{.Label}}</label>
        <textarea id="f-{{.Name}}" name="{{.Name}}"{{if .Required}} required{{end}}>{{.Value}}</textarea>
        {{if .Help}}<div class="help">{{.Help}}</div>{{end}}

      {{else if eq (printf "%s" .Type) "number"}}
        <label for="f-{{.Name}}">{{.Label}}</label>
        <input type="number" id="f-{{.Name}}" name="{{.Name}}" value="{{.Value}}"{{if .Required}} required{{end}} step="any">
        {{if .Help}}<div class="help">{{.Help}}</div>{{end}}

      {{else if eq (printf "%s" .Type) "date"}}
        <label for="f-{{.Name}}">{{.Label}}</label>
        <input type="date" id="f-{{.Name}}" name="{{.Name}}" value="{{.Value}}"{{if .Required}} required{{end}}>
        {{if .Help}}<div class="help">{{.Help}}</div>{{end}}

      {{else if or (eq (printf "%s" .Type) "select") (eq (printf "%s" .Type) "relation")}}
        <label for="f-{{.Name}}">{{.Label}}</label>
        <select id="f-{{.Name}}" name="{{.Name}}"{{if .Required}} required{{end}}>
          <option value="">— не выбрано —</option>
          {{$val := .Value}}
          {{range .Options}}<option value="{{.Value}}"{{if eq .Value $val}} selected{{end}}>{{.Label}}</option>{{end}}
        </select>
        {{if .Help}}<div class="help">{{.Help}}</div>{{end}}

      {{else if eq (printf "%s" .Type) "file"}}
        <label>{{.Label}}</label>
        <div class="filezone">
          <input type="file" name="{{.Name}}">
          <input type="hidden" name="{{.Name}}_key" value="{{.Value}}">
          {{if .FileURL}}<span class="filepreview"><img class="thumb" src="{{.FileURL}}" alt=""> Текущий файл</span>{{end}}
        </div>
        {{if .Help}}<div class="help">{{.Help}}</div>{{end}}

      {{else}}
        <label for="f-{{.Name}}">{{.Label}}</label>
        <input type="text" id="f-{{.Name}}" name="{{.Name}}" value="{{.Value}}"{{if .Required}} required{{end}}>
        {{if .Help}}<div class="help">{{.Help}}</div>{{end}}
      {{end}}
    </div>
  {{end}}
{{end}}
```

## `internal/admin/templates/layout.html`

```html
{{define "head"}}
<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} · Панель управления</title>
<!-- htmx: интерактив без перезагрузки страницы (раздаётся web-сервером приложения). -->
<script src="/static/htmx.min.js"></script>
<style>
:root{
  --background:#FFFFFF; --foreground:#09090B;
  --muted:#F4F4F5; --muted-foreground:#71717A;
  --border:#E4E4E7; --subtle:#FAFAFA;
  --primary:#18181B; --primary-foreground:#FAFAFA;
  --secondary:#F4F4F5; --secondary-foreground:#18181B;
  --accent:#F4F4F5; --ring:#18181B;
  --destructive:#DC2626; --destructive-foreground:#FFFFFF; --destructive-subtle:#FEF2F2;
  --success:#16A34A; --success-foreground:#166534; --success-subtle:#DCFCE7;
  --warning:#D97706; --warning-foreground:#92400E; --warning-subtle:#FEF3C7;
  --info:#2563EB; --info-foreground:#1E40AF; --info-subtle:#DBEAFE;
  --radius:6px; --radius-lg:8px; --radius-sm:4px;
}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:Inter,system-ui,-apple-system,"Segoe UI",sans-serif;color:var(--foreground);background:var(--background);font-size:14px;line-height:1.45}
a{color:inherit;text-decoration:none}
.layout{display:flex;min-height:100vh}
/* --- Sidebar --- */
.sidebar{width:248px;flex-shrink:0;border-right:1px solid var(--border);background:var(--subtle);display:flex;flex-direction:column;padding:16px 12px}
.sidebar__brand{display:flex;align-items:center;gap:8px;font-weight:600;font-size:15px;padding:8px 12px 16px}
.sidebar__nav{display:flex;flex-direction:column;gap:2px;flex:1}
.navitem{display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:var(--radius);color:var(--muted-foreground);font-weight:500}
.navitem svg{flex:none;opacity:.85}
.navitem:hover{background:var(--accent);color:var(--foreground)}
.navitem--active{background:var(--secondary);color:var(--secondary-foreground)}
.navitem--active svg{opacity:1}
.sidebar__user{border-top:1px solid var(--border);padding:12px 8px 0;display:flex;align-items:center;justify-content:space-between;gap:8px;color:var(--muted-foreground);font-size:13px}
.sidebar__user form{margin:0}
.sidebar__user .btn{display:inline-flex;align-items:center;gap:6px}
/* --- Content --- */
.content{flex:1;min-width:0;padding:28px 32px;max-width:100%}
.topbar{display:flex;align-items:center;justify-content:space-between;margin-bottom:20px}
.topbar h1{font-size:22px;font-weight:600}
.crumbs{color:var(--muted-foreground);font-size:13px;margin-bottom:6px}
/* --- Buttons --- */
.btn{display:inline-flex;align-items:center;gap:6px;height:36px;padding:0 14px;border-radius:var(--radius);font-weight:500;font-size:14px;border:1px solid transparent;cursor:pointer;background:none;font-family:inherit}
.btn--default{background:var(--primary);color:var(--primary-foreground)}
.btn--default:hover{background:#000}
.btn--outline{border-color:var(--border);background:var(--background);color:var(--foreground)}
.btn--outline:hover{background:var(--accent)}
.btn--ghost{color:var(--foreground)}
.btn--ghost:hover{background:var(--accent)}
.btn--destructive{background:var(--destructive);color:var(--destructive-foreground)}
.btn--destructive:hover{background:#b91c1c}
.btn--sm{height:32px;padding:0 10px;font-size:13px}
.btn--icon{height:32px;width:32px;padding:0;justify-content:center;color:var(--muted-foreground)}
.btn--icon:hover{background:var(--accent);color:var(--foreground)}
/* --- Toolbar --- */
.toolbar{display:flex;align-items:center;gap:8px;margin-bottom:16px;flex-wrap:wrap}
.search{flex:1;min-width:220px;position:relative;display:flex;align-items:center}
.search .search__icon{position:absolute;left:11px;color:var(--muted-foreground);pointer-events:none;display:inline-flex}
.search input{width:100%;height:36px;padding:0 12px 0 34px;border:1px solid var(--border);border-radius:var(--radius);font-size:14px;font-family:inherit;background:var(--background)}
.btn svg{flex:none}
.select{height:36px;padding:0 10px;border:1px solid var(--border);border-radius:var(--radius);background-color:var(--background);font-size:14px;font-family:inherit;color:var(--foreground)}
input:focus,select:focus,textarea:focus{outline:none;border-color:var(--ring);box-shadow:0 0 0 3px rgba(24,24,27,.08)}
/* Своя стрелка у select (нативная прижата к краю) — chevron-down с отступом справа */
select{
  appearance:none;-webkit-appearance:none;-moz-appearance:none;
  background-image:url('data:image/svg+xml,%3Csvg%20xmlns=%22http://www.w3.org/2000/svg%22%20width=%2216%22%20height=%2216%22%20viewBox=%220%200%2024%2024%22%20fill=%22none%22%20stroke=%22%2371717A%22%20stroke-width=%222.25%22%20stroke-linecap=%22round%22%20stroke-linejoin=%22round%22%3E%3Cpath%20d=%22m6%209%206%206%206-6%22/%3E%3C/svg%3E');
  background-repeat:no-repeat;background-position:right 11px center;background-size:16px;
  padding-right:34px !important;cursor:pointer;
}
/* --- Table --- */
.card{border:1px solid var(--border);border-radius:var(--radius-lg);overflow:hidden;background:var(--background)}
.card:has(table){overflow-x:auto}
table{width:100%;border-collapse:collapse}
thead th{text-align:left;font-weight:500;font-size:12px;color:var(--muted-foreground);padding:10px 14px;border-bottom:1px solid var(--border);background:var(--subtle);white-space:nowrap}
thead th a{display:inline-flex;align-items:center;gap:4px}
tbody td{padding:11px 14px;border-bottom:1px solid var(--border);vertical-align:middle}
tbody tr:last-child td{border-bottom:none}
tbody tr:hover{background:var(--subtle)}
.checkcol{width:36px;text-align:center}
.actions{white-space:nowrap;text-align:right;width:1%}
.thumb{width:28px;height:28px;border-radius:var(--radius-sm);object-fit:cover;border:1px solid var(--border);background:var(--muted)}
.tick{color:var(--success)}
.tick--off{color:var(--muted-foreground)}
/* --- Badge --- */
.badge{display:inline-flex;align-items:center;padding:2px 8px;border-radius:999px;font-size:12px;font-weight:500;border:1px solid var(--border);background:var(--secondary);color:var(--secondary-foreground)}
.badge--success{background:var(--success-subtle);color:var(--success-foreground);border-color:transparent}
.badge--warning{background:var(--warning-subtle);color:var(--warning-foreground);border-color:transparent}
.badge--destructive{background:var(--destructive-subtle);color:var(--destructive);border-color:transparent}
.badge--info{background:var(--info-subtle);color:var(--info-foreground);border-color:transparent}
/* --- Bulk bar --- */
.bulkbar{display:flex;align-items:center;justify-content:space-between;padding:8px 14px;border:1px solid var(--border);border-radius:var(--radius);background:var(--subtle);margin-bottom:12px}
.bulkbar[hidden]{display:none}
/* --- Pagination --- */
.pagination{display:flex;align-items:center;justify-content:space-between;padding:12px 4px;color:var(--muted-foreground);font-size:13px}
.pagination__pages{display:flex;gap:4px;align-items:center}
/* --- Empty --- */
.empty{text-align:center;padding:64px 24px;color:var(--muted-foreground)}
.empty h2{font-size:16px;color:var(--foreground);margin-bottom:6px;font-weight:600}
.empty p{margin-bottom:16px}
/* --- Form --- */
.formwrap{max-width:640px}
.field{margin-bottom:18px}
.field label{display:block;font-weight:500;margin-bottom:6px}
.field .help{color:var(--muted-foreground);font-size:13px;margin-top:5px}
.field input[type=text],.field input[type=number],.field input[type=date],.field select,.field textarea{
  width:100%;padding:9px 12px;border:1px solid var(--border);border-radius:var(--radius);font-size:14px;font-family:inherit;background-color:var(--background)}
.field textarea{min-height:96px;resize:vertical}
.field--error input,.field--error textarea,.field--error select{border-color:var(--destructive)}
.field__error{color:var(--destructive);font-size:13px;margin-top:6px}
.row2{display:flex;gap:16px}
.row2>.field{flex:1}
/* switch — строка «подпись + описание слева, тумблер справа» (shadcn switch-with-label) */
.field label.switchrow,.switchrow{display:flex;align-items:center;justify-content:space-between;gap:16px;cursor:pointer;
  border:1px solid var(--border);border-radius:var(--radius);padding:12px 14px;margin-bottom:0}
.switchrow:hover{background:var(--subtle)}
.switchrow__text{display:flex;flex-direction:column;gap:2px;min-width:0}
.switchrow__label{font-weight:500}
.switchrow__help{font-size:13px;color:var(--muted-foreground)}
.switchrow input{position:absolute;opacity:0;width:0;height:0;pointer-events:none}
.switch__track{display:inline-block;flex:none;width:40px;height:22px;border-radius:999px;background:var(--border);position:relative;transition:background .15s}
.switch__thumb{position:absolute;top:2px;left:2px;width:18px;height:18px;border-radius:50%;background:#fff;transition:left .15s;box-shadow:0 1px 2px rgba(0,0,0,.2)}
.switchrow input:checked+.switch__track{background:var(--primary)}
.switchrow input:checked+.switch__track .switch__thumb{left:20px}
/* file */
.filezone{border:1px dashed var(--border);border-radius:var(--radius);padding:16px;background:var(--subtle);display:flex;align-items:center;gap:12px}
.filezone input{font-family:inherit;font-size:13px}
.filepreview{display:flex;align-items:center;gap:10px;font-size:13px;color:var(--muted-foreground)}
.formactions{display:flex;justify-content:flex-end;gap:10px;border-top:1px solid var(--border);padding-top:18px;margin-top:8px}
/* modal (htmx-форма создания/редактирования) */
.modal-overlay{position:fixed;inset:0;background:rgba(0,0,0,.4);display:flex;align-items:flex-start;justify-content:center;padding:48px 16px;z-index:50;overflow:auto}
.modal{background:var(--background);border:1px solid var(--border);border-radius:var(--radius-lg);box-shadow:0 10px 40px rgba(0,0,0,.18);width:100%;max-width:640px}
.modal__head{display:flex;align-items:center;justify-content:space-between;padding:18px 22px;border-bottom:1px solid var(--border)}
.modal__head h2{font-size:17px;font-weight:600}
.modal__body{padding:20px 22px}
.htmx-indicator{opacity:0;transition:opacity .15s}
.htmx-request .htmx-indicator,.htmx-request.htmx-indicator{opacity:1}

/* --- Адаптив --- */
@media (max-width:1024px){
  .content{padding:20px 18px}
}
@media (max-width:768px){
  /* Сайдбар → горизонтальная полоса сверху; страница в колонку */
  .layout{flex-direction:column}
  .sidebar{width:auto;flex-direction:row;align-items:center;gap:6px;padding:10px 12px;
    border-right:none;border-bottom:1px solid var(--border);overflow-x:auto}
  .sidebar__brand{padding:0 8px 0 0;white-space:nowrap}
  .sidebar__brand span{display:none}            /* на мобиле бренд — только иконка */
  .sidebar__nav{flex-direction:row;gap:4px;flex:1}
  .navitem{padding:8px 10px;white-space:nowrap}
  .sidebar__user{border-top:none;padding:0;margin-left:auto;flex:none}
  .sidebar__user>span{display:none}             /* «Администратор» прячем, оставляем «Выйти» */
  .content{padding:16px 14px}
  .topbar{flex-wrap:wrap;gap:10px}
  .topbar h1{font-size:19px}
  /* Тулбар переносится, поиск на всю ширину */
  .toolbar{gap:8px}
  .search{min-width:100%;order:-1}
  /* Модалка почти на всю ширину */
  .modal-overlay{padding:16px 10px}
  .row2{flex-direction:column;gap:0}
  .pagination{flex-wrap:wrap;gap:8px}
}
</style>
</head>
<body>
<div class="layout">
  <aside class="sidebar">
    <div class="sidebar__brand">{{icon "layout-dashboard" 18}} <span>Панель управления</span></div>
    <nav class="sidebar__nav" id="admin-nav">
      {{range .Menu}}
        <a class="navitem{{if .Active}} navitem--active{{end}}"
           href="/admin/{{.Name}}"
           hx-get="/admin/{{.Name}}" hx-target="#admin-content" hx-swap="innerHTML" hx-push-url="true">{{icon .Icon 17}} <span>{{.Title}}</span></a>
      {{end}}
    </nav>
    <div class="sidebar__user">
      <span>Администратор</span>
      <button class="btn btn--ghost btn--sm" hx-post="/admin/logout" title="Выйти">{{icon "log-out" 15}} <span>Выйти</span></button>
    </div>
  </aside>
  <main class="content">
    <div id="admin-content">
{{end}}

{{define "foot"}}
    </div>
  </main>
</div>
<!-- Контейнер модалки форм: htmx грузит сюда форму создания/редактирования. -->
<div id="modal"></div>
<script>
  // Закрытие модалки по клику на фон/крестик/«Отмена» (элементы с data-close).
  document.addEventListener('click', (e) => {
    if (e.target.dataset && e.target.dataset.close !== undefined) {
      document.getElementById('modal').innerHTML = '';
    }
  });
  // После успешного создания/редактирования: модалка уже очищена пустым ответом,
  // осталось перезагрузить секцию списка БЕЗ перезагрузки страницы.
  document.body.addEventListener('refreshList', () => {
    const lb = document.getElementById('listbody');
    if (lb && lb.dataset.resource) {
      htmx.ajax('GET', '/admin/' + lb.dataset.resource, { target: '#listbody', swap: 'innerHTML' });
    }
  });
  // Ошибки мутаций (например, 422 «нельзя удалить последнего админа») — показать текст.
  document.body.addEventListener('htmx:responseError', (e) => {
    const x = e.detail.xhr;
    if (x && x.status >= 400 && x.status < 500) { alert((x.responseText || '').trim() || 'Ошибка'); }
  });
  // Подсветка активного раздела при HTMX-навигации по сайдбару (контент подменяется,
  // сам сайдбар нет — переключаем класс на клиенте, без перезагрузки).
  document.getElementById('admin-nav').addEventListener('click', (e) => {
    const item = e.target.closest('.navitem');
    if (!item) return;
    document.querySelectorAll('#admin-nav .navitem').forEach(n => n.classList.remove('navitem--active'));
    item.classList.add('navitem--active');
  });
</script>
</body>
</html>
{{end}}
```

## `internal/admin/templates/list.html`

```html
{{/* list — полная страница (прямой заход по URL раздела). */}}
{{define "list"}}
{{template "head" .}}
  {{template "content" .}}
{{template "foot" .}}
{{end}}

{{/* content — контент раздела (топбар + список) БЕЗ каркаса. Возвращается при
     HTMX-навигации по сайдбару (подмена #admin-content), без перезагрузки страницы. */}}
{{define "content"}}
  <div class="topbar">
    <h1>{{.Title}}</h1>
    <button class="btn btn--default"
            hx-get="/admin/{{.Resource}}/new" hx-target="#modal" hx-swap="innerHTML">{{icon "plus" 16}} <span>Создать</span></button>
  </div>

  <div id="listbody" data-resource="{{.Resource}}">
    {{template "listbody" .}}
  </div>
{{end}}


{{define "listbody"}}
  {{/* Вся секция списка подменяется целиком при поиске/фильтре/сортировке/пагинации
       (hx-get → #listbody). Внутри — тулбар, таблица, пагинация. */}}
  <form class="toolbar"
        hx-get="/admin/{{.Resource}}" hx-target="#listbody" hx-swap="innerHTML"
        hx-trigger="submit, change, keyup from:input[name='q'] changed delay:300ms"
        hx-push-url="true">
    <div class="search">
      <span class="search__icon">{{icon "search" 15}}</span>
      <input type="text" name="q" value="{{.Search}}" placeholder="Поиск по разделу…" autocomplete="off">
    </div>
    {{range .Filters}}
      <select class="select" name="f_{{.Name}}">
        <option value="">{{.Label}}: все</option>
        {{$sel := .Selected}}
        {{range .Options}}<option value="{{.Value}}"{{if eq .Value $sel}} selected{{end}}>{{.Label}}</option>{{end}}
      </select>
    {{end}}
    <input type="hidden" name="sort" value="{{.SortBy}}">
    <input type="hidden" name="dir" value="{{.SortDir}}">
    <span class="htmx-indicator" style="color:var(--muted-foreground);font-size:13px">…</span>
    <a class="btn btn--outline" href="/admin/{{.Resource}}/export?q={{.Search}}">{{icon "download" 15}} <span>Экспорт CSV</span></a>
  </form>

  <div class="bulkbar" id="bulkbar" hidden>
    <span>Выбрано: <b id="bulkcount">0</b></span>
    <button class="btn btn--destructive btn--sm" type="button" onclick="bulkDelete('{{.Resource}}')">Удалить выбранные</button>
  </div>

  {{if .Rows}}
  <div class="card">
    <table>
      <thead>
        <tr>
          <th class="checkcol"><input type="checkbox" onclick="toggleAll(this)"></th>
          {{$res := .Resource}}{{$dir := .SortDir}}{{$sort := .SortBy}}{{$q := .Search}}
          {{range .Columns}}
            <th>
              {{if .Sortable}}
                <a href="#" style="cursor:pointer"
                   hx-get="/admin/{{$res}}?sort={{.Name}}&dir={{if and (eq $sort .Name) (eq $dir "asc")}}desc{{else}}asc{{end}}&q={{$q}}"
                   hx-target="#listbody" hx-swap="innerHTML" hx-push-url="true">
                  {{.Label}}{{if .SortedAsc}} {{icon "arrow-up" 13}}{{else if .SortedDesc}} {{icon "arrow-down" 13}}{{end}}
                </a>
              {{else}}{{.Label}}{{end}}
            </th>
          {{end}}
          <th class="actions">Действия</th>
        </tr>
      </thead>
      <tbody>
        {{range .Rows}}
        <tr id="row-{{$res}}-{{.ID}}">
          <td class="checkcol"><input type="checkbox" class="rowcheck" value="{{.ID}}" onclick="updateBulk()"></td>
          {{range .Cells}}
            <td>
              {{if .IsBool}}
                {{if .Bool}}<span class="tick">{{icon "check" 16}}</span>{{else}}<span class="tick--off">{{icon "minus" 16}}</span>{{end}}
              {{else if .IsFile}}
                {{if .FileURL}}<img class="thumb" src="{{.FileURL}}" alt="{{.Display}}">{{else}}<span class="tick--off">{{icon "minus" 16}}</span>{{end}}
              {{else if .Tone}}
                <span class="badge badge--{{.Tone}}">{{.Display}}</span>
              {{else}}{{.Display}}{{end}}
            </td>
          {{end}}
          <td class="actions">
            <button class="btn btn--icon" title="Изменить"
                    hx-get="/admin/{{$res}}/{{.ID}}/edit" hx-target="#modal" hx-swap="innerHTML">{{icon "pencil" 16}}</button>
            <button class="btn btn--icon" title="Удалить"
                    hx-delete="/admin/{{$res}}/{{.ID}}"
                    hx-target="#listbody" hx-swap="innerHTML"
                    hx-confirm="Удалить запись? Действие необратимо.">{{icon "trash" 16}}</button>
          </td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>

  <div class="pagination">
    <span>Показано {{.From}}–{{.To}} из {{.Total}}</span>
    <div class="pagination__pages">
      {{$res := .Resource}}{{$q := .Search}}{{$sort := .SortBy}}{{$dir := .SortDir}}
      {{if .HasPrev}}<button class="btn btn--outline btn--sm"
        hx-get="/admin/{{$res}}?page={{.PrevPage}}&q={{$q}}&sort={{$sort}}&dir={{$dir}}"
        hx-target="#listbody" hx-swap="innerHTML" hx-push-url="true">{{icon "chevron-left" 14}} <span>Назад</span></button>{{end}}
      <span>{{.Page}} / {{.Pages}}</span>
      {{if .HasNext}}<button class="btn btn--outline btn--sm"
        hx-get="/admin/{{$res}}?page={{.NextPage}}&q={{$q}}&sort={{$sort}}&dir={{$dir}}"
        hx-target="#listbody" hx-swap="innerHTML" hx-push-url="true"><span>Вперёд</span> {{icon "chevron-right" 14}}</button>{{end}}
    </div>
  </div>
  {{else}}
  <div class="card">
    <div class="empty">
      <h2>Здесь пока ничего нет</h2>
      <p>Создайте первую запись, чтобы она появилась в этом разделе.</p>
      <button class="btn btn--default"
              hx-get="/admin/{{.Resource}}/new" hx-target="#modal" hx-swap="innerHTML">{{icon "plus" 16}} <span>Создать</span></button>
    </div>
  </div>
  {{end}}

  <script>
    function toggleAll(box){document.querySelectorAll('.rowcheck').forEach(c=>c.checked=box.checked);updateBulk();}
    function updateBulk(){
      const sel=[...document.querySelectorAll('.rowcheck:checked')];
      const bc=document.getElementById('bulkcount'), bb=document.getElementById('bulkbar');
      if(bc) bc.textContent=sel.length;
      if(bb) bb.hidden=sel.length===0;
    }
    async function bulkDelete(res){
      const sel=[...document.querySelectorAll('.rowcheck:checked')].map(c=>c.value);
      if(!sel.length)return;
      if(!confirm('Удалить выбранные ('+sel.length+')? Действие необратимо.'))return;
      for(const id of sel){ await fetch('/admin/'+res+'/'+id,{method:'DELETE',headers:{'HX-Request':'true'}}); }
      // перезагрузим секцию списка без перезагрузки страницы
      htmx.ajax('GET','/admin/'+res,{target:'#listbody',swap:'innerHTML'});
    }
  </script>
{{end}}
```

## `internal/admin/templates/login.html`

```html
{{/* login — вход в админку. Светлый shadcn-стиль, единый с самой админкой (та же
     zinc-палитра). Отдельная страница (не часть приложения), но визуально согласована
     с внутренними экранами панели управления. */}}
{{define "login"}}
<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Вход · Панель управления</title>
<script src="/static/htmx.min.js"></script>
<style>
:root{
  --background:#FFFFFF; --foreground:#09090B;
  --muted:#F4F4F5; --muted-foreground:#71717A;
  --border:#E4E4E7; --subtle:#FAFAFA;
  --primary:#18181B; --primary-foreground:#FAFAFA;
  --ring:#18181B; --destructive:#DC2626; --destructive-subtle:#FEF2F2;
  --radius:6px; --radius-lg:8px;
}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:Inter,system-ui,-apple-system,"Segoe UI",sans-serif;color:var(--foreground);
  background:var(--subtle);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:24px}
.card{width:100%;max-width:380px;background:var(--background);border:1px solid var(--border);
  border-radius:var(--radius-lg);padding:32px;box-shadow:0 1px 3px rgba(0,0,0,.06),0 8px 24px rgba(0,0,0,.04)}
.brand{display:inline-flex;align-items:center;gap:8px;font-weight:600;font-size:15px;margin-bottom:20px}
.badge{display:inline-flex;align-items:center;gap:6px;font-size:12px;color:var(--muted-foreground);
  border:1px solid var(--border);border-radius:999px;padding:4px 10px;margin-bottom:16px}
.dot{width:6px;height:6px;border-radius:50%;background:#16A34A}
h1{font-size:20px;font-weight:600;margin-bottom:6px}
.sub{color:var(--muted-foreground);font-size:14px;margin-bottom:22px}
label{display:block;font-size:13px;font-weight:500;margin-bottom:6px;margin-top:16px}
label:first-of-type{margin-top:0}
input{width:100%;height:40px;padding:0 12px;background:var(--background);border:1px solid var(--border);
  border-radius:var(--radius);color:var(--foreground);font-size:14px;font-family:inherit}
input:focus{outline:none;border-color:var(--ring);box-shadow:0 0 0 3px rgba(24,24,27,.08)}
.btn{width:100%;height:40px;margin-top:24px;background:var(--primary);color:var(--primary-foreground);
  border:none;border-radius:var(--radius);font-size:14px;font-weight:500;cursor:pointer;font-family:inherit}
.btn:hover{background:#000}
.error{background:var(--destructive-subtle);color:var(--destructive);border:1px solid var(--destructive);
  border-radius:var(--radius);padding:10px 13px;font-size:13px;margin-bottom:18px}
.foot{margin-top:22px;text-align:center;font-size:13px}
.foot a{color:var(--muted-foreground);text-decoration:none}
.foot a:hover{color:var(--foreground)}
</style>
</head>
<body>
  <div class="card">
    <span class="badge"><span class="dot"></span> Панель управления</span>
    <h1>Вход</h1>
    <p class="sub">Доступ только для администраторов приложения.</p>
    {{if .Err}}<div class="error">{{.Err}}</div>{{end}}
    <form hx-post="/admin/login" hx-target="this" hx-swap="none">
      <label for="email">Email</label>
      <input type="email" id="email" name="email" value="{{.Email}}" autocomplete="username" required autofocus>
      <label for="password">Пароль</label>
      <input type="password" id="password" name="password" autocomplete="current-password" required>
      <button class="btn" type="submit">Войти</button>
    </form>
    <div class="foot"><a href="/">← Вернуться в приложение</a></div>
  </div>
</body>
</html>
{{end}}
```

## `internal/admin/view.go`

```go
package admin

import (
	"context"
	"html/template"
)

// adminFuncs — функции, доступные в админ-шаблонах.
var adminFuncs = template.FuncMap{
	// dict собирает map для передачи нескольких значений во вложенный шаблон.
	"dict": func(pairs ...any) map[string]any {
		m := map[string]any{}
		for i := 0; i+1 < len(pairs); i += 2 {
			k, _ := pairs[i].(string)
			m[k] = pairs[i+1]
		}
		return m
	},
	// icon рендерит inline-SVG Lucide-иконку: {{icon "pencil" 16}}.
	"icon": icon,
}

// menuItem — пункт бокового меню (раздел сущности).
type menuItem struct {
	Name   string
	Title  string
	Icon   string
	Active bool
}

// menu строит пункты меню из реестра, отмечая активный раздел.
func (s *Server) menu(active string) []menuItem {
	items := make([]menuItem, 0)
	for _, h := range s.reg.All() {
		items = append(items, menuItem{Name: h.Name(), Title: h.Title(), Icon: h.Icon(), Active: h.Name() == active})
	}
	return items
}

// column — заголовок колонки списка (с состоянием сортировки).
type column struct {
	Name       string
	Label      string
	Sortable   bool
	SortedAsc  bool
	SortedDesc bool
}

// filterView — выпадающий фильтр над таблицей (по select/badge-полю).
type filterView struct {
	Name     string
	Label    string
	Options  []Option
	Selected string
}

// listView — данные экрана списка.
type listView struct {
	Title    string
	Resource string
	Menu     []menuItem
	Columns  []column
	Filters  []filterView
	Rows     []rowView
	Search   string
	SortBy   string
	SortDir  string
	Page     int
	Pages    int
	Total    int
	From     int
	To       int
	HasPrev  bool
	HasNext  bool
	PrevPage int
	NextPage int
}

// rowView — строка таблицы: id и ячейки в порядке колонок.
type rowView struct {
	ID    string
	Cells []cellView
}

// cellView — ячейка таблицы (с уже подготовленным отображением).
type cellView struct {
	Display string
	Tone    string
	IsFile  bool
	FileURL string
	IsBool  bool
	Bool    bool
}

// listView собирает вью-модель списка из окна пагинации и запроса.
func (s *Server) listView(ctx context.Context, h ResourceHandle, win pageWindow, q ListQuery) listView {
	var cols []column
	var filters []filterView
	for _, f := range h.Fields() {
		if f.InList {
			cols = append(cols, column{
				Name: f.Name, Label: f.Label, Sortable: f.Sortable,
				SortedAsc:  q.SortBy == f.Name && q.SortDir == "asc",
				SortedDesc: q.SortBy == f.Name && q.SortDir == "desc",
			})
		}
		// Фильтры — по полям с фиксированными вариантами (select/badge).
		if (f.Type == FieldSelect || f.Type == FieldBadge) && len(f.Options) > 0 {
			filters = append(filters, filterView{
				Name: f.Name, Label: f.Label, Options: f.Options, Selected: q.Filters[f.Name],
			})
		}
	}

	rows := make([]rowView, 0, len(win.Records))
	for _, rec := range win.Records {
		cells := make([]cellView, 0, len(cols))
		for _, col := range cols {
			c := rec.Cells[col.Name]
			cv := cellView{Display: c.Display, Tone: c.Tone, IsFile: c.IsFile}
			if cv.Display == "" {
				cv.Display = c.Value
			}
			f, _ := fieldByName(h, col.Name)
			if f.Type == FieldBool {
				cv.IsBool = true
				cv.Bool = c.Value == "true"
			}
			if c.IsFile && c.Value != "" && s.files != nil {
				if url, err := s.files.URL(ctx, c.Value); err == nil {
					cv.FileURL = url
				}
			}
			cells = append(cells, cv)
		}
		rows = append(rows, rowView{ID: rec.ID, Cells: cells})
	}

	from := 0
	if win.Total > 0 {
		from = (win.Page-1)*win.PageSize + 1
	}
	to := from + len(win.Records) - 1
	if to < 0 {
		to = 0
	}
	return listView{
		Title:    h.Title(),
		Resource: h.Name(),
		Menu:     s.menu(h.Name()),
		Columns:  cols,
		Filters:  filters,
		Rows:     rows,
		Search:   q.Search,
		SortBy:   q.SortBy,
		SortDir:  q.SortDir,
		Page:     win.Page,
		Pages:    win.Pages,
		Total:    win.Total,
		From:     from,
		To:       to,
		HasPrev:  win.Page > 1,
		HasNext:  win.Page < win.Pages,
		PrevPage: win.Page - 1,
		NextPage: win.Page + 1,
	}
}

// fieldView — поле формы с текущим значением и (для связей) подгруженными вариантами.
type fieldView struct {
	Name     string
	Label    string
	Type     FieldType
	Required bool
	Help     string
	Value    string
	Options  []Option
	FileURL  string // для file-поля: ссылка на текущий загруженный файл
}

// formView — данные экрана формы (создание/редактирование).
type formView struct {
	Title    string
	Resource string
	Menu     []menuItem
	ID       string // пусто → создание; иначе редактирование
	Fields   []fieldView
	Err      string
}

// formView собирает вью-модель формы, подгружая варианты для связей и URL файлов.
func (s *Server) formView(ctx context.Context, h ResourceHandle, id string, values FormValues, errMsg string) formView {
	fields := make([]fieldView, 0, len(h.Fields()))
	for _, f := range h.Fields() {
		fv := fieldView{
			Name: f.Name, Label: f.Label, Type: f.Type,
			Required: f.Required, Help: f.Help, Value: values[f.Name],
			Options: f.Options,
		}
		if f.Type == FieldRelation && f.RelationOptions != nil {
			if opts, err := f.RelationOptions(ctx); err == nil {
				fv.Options = opts
			}
		}
		if f.Type == FieldFile && fv.Value != "" && s.files != nil {
			if url, err := s.files.URL(ctx, fv.Value); err == nil {
				fv.FileURL = url
			}
		}
		fields = append(fields, fv)
	}
	return formView{
		Title:    h.Title(),
		Resource: h.Name(),
		Menu:     s.menu(h.Name()),
		ID:       id,
		Fields:   fields,
		Err:      errMsg,
	}
}
```

## `internal/adminres/note.go`

```go
// Package adminres регистрирует конкретные сущности приложения в админ-реестре.
// Отделён от пакета admin (generic-ядро не знает о Note/User) — здесь стык: дескриптор
// связывает поля сущности с её репозиторием. ОБРАЗЕЦ ДЛЯ ГЕНЕРАЦИИ: новая сущность
// под админку описывается такой же функцией RegisterX(reg, repo).
package adminres

import (
	"context"
	"strconv"
	"time"

	"github.com/chudno/zerovibe/internal/admin"
	"github.com/chudno/zerovibe/internal/domain"
)

// formatDate переводит YYYY-MM-DD в DD.MM.YYYY для показа в списке. Некорректную
// строку отдаёт как есть (на всякий случай).
func formatDate(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return t.Format("02.01.2006")
}

// NoteAdminRepo — что админке нужно от репозитория заметок (admin-CRUD над всеми).
type NoteAdminRepo interface {
	ListAll(ctx context.Context) ([]domain.Note, error)
	GetByID(ctx context.Context, id int64) (domain.Note, error)
	Create(ctx context.Context, n domain.Note) (domain.Note, error)
	UpdateAny(ctx context.Context, n domain.Note) error
	DeleteAny(ctx context.Context, id int64) error
}

// ownerOptions — источник вариантов для связи «Владелец» (список пользователей).
type ownerOptions func(ctx context.Context) ([]admin.Option, error)

// RegisterNote регистрирует сущность «Заметки» в админ-реестре. Это ЭТАЛОН: показывает
// все типы участников — поля разных типов, связь (Владелец → пользователь), список с
// сортировкой/поиском, полный CRUD. Новая сущность повторяет эту структуру.
func RegisterNote(reg *admin.Registry, repo NoteAdminRepo, owners ownerOptions) {
	admin.Register(reg, admin.Resource[domain.Note]{
		Name:  "notes",
		Title: "Заметки",
		Icon:  "file-text",
		Fields: []admin.Field{
			{Name: "title", Label: "Заголовок", Type: admin.FieldText, Required: true, InList: true, Sortable: true},
			{Name: "body", Label: "Текст", Type: admin.FieldTextarea, InList: false, Help: "Содержимое заметки"},
			{
				Name: "owner_id", Label: "Владелец", Type: admin.FieldRelation, Required: true, InList: true,
				RelationOptions: func(ctx context.Context) ([]admin.Option, error) { return owners(ctx) },
			},
			// FieldDate → в форме нативный календарь (input type=date).
			{Name: "due_date", Label: "Срок", Type: admin.FieldDate, InList: true, Sortable: true, Help: "Необязательно"},
			{Name: "created_at", Label: "Создана", Type: admin.FieldDate, InList: true, Sortable: true},
		},

		// Row — как заметка выглядит строкой списка.
		Row: func(n domain.Note) admin.Record {
			due := admin.Cell{}
			if n.DueDate != "" {
				due = admin.Cell{Value: n.DueDate, Display: formatDate(n.DueDate)}
			}
			return admin.Record{
				ID: strconv.FormatInt(n.ID, 10),
				Cells: map[string]admin.Cell{
					"title":      {Value: n.Title, Display: n.Title},
					"owner_id":   {Value: strconv.FormatInt(n.OwnerID, 10), Display: "#" + strconv.FormatInt(n.OwnerID, 10)},
					"due_date":   due,
					"created_at": {Value: n.CreatedAt.Format("2006-01-02"), Display: n.CreatedAt.Format("02.01.2006")},
				},
			}
		},

		List: func(ctx context.Context) ([]domain.Note, error) { return repo.ListAll(ctx) },

		// Get — значения для формы редактирования. Дата — в формате YYYY-MM-DD (его
		// ждёт input type=date).
		Get: func(ctx context.Context, id string) (admin.FormValues, error) {
			nid, _ := strconv.ParseInt(id, 10, 64)
			n, err := repo.GetByID(ctx, nid)
			if err != nil {
				return nil, err
			}
			return admin.FormValues{
				"title":    n.Title,
				"body":     n.Body,
				"owner_id": strconv.FormatInt(n.OwnerID, 10),
				"due_date": n.DueDate,
			}, nil
		},

		// Create — валидация через доменный конструктор, затем сохранение.
		Create: func(ctx context.Context, v admin.FormValues, _ map[string]admin.FileInput) error {
			n, err := domain.NewNote(v["title"], v["body"], v["due_date"])
			if err != nil {
				return err
			}
			n.OwnerID, _ = strconv.ParseInt(v["owner_id"], 10, 64)
			_, err = repo.Create(ctx, n)
			return err
		},

		Update: func(ctx context.Context, id string, v admin.FormValues, _ map[string]admin.FileInput) error {
			n, err := domain.NewNote(v["title"], v["body"], v["due_date"])
			if err != nil {
				return err
			}
			n.ID, _ = strconv.ParseInt(id, 10, 64)
			return repo.UpdateAny(ctx, n)
		},

		Delete: func(ctx context.Context, id string) error {
			nid, _ := strconv.ParseInt(id, 10, 64)
			return repo.DeleteAny(ctx, nid)
		},
	})
}
```

## `internal/adminres/user.go`

```go
package adminres

import (
	"context"
	"strconv"

	"github.com/chudno/zerovibe/internal/admin"
	"github.com/chudno/zerovibe/internal/domain"
)

// UserAdminRepo — что админке нужно от репозитория пользователей.
type UserAdminRepo interface {
	ListAll(ctx context.Context) ([]domain.User, error)
	ByID(ctx context.Context, id int64) (domain.User, error)
	Create(ctx context.Context, u domain.User) (domain.User, error)
	UpdateRoleAndEmail(ctx context.Context, id int64, email string, role domain.Role, verified bool) error
	DeleteUser(ctx context.Context, id int64) error
}

// Hasher — хеширование пароля (bcrypt из usecase). Пароль НИКОГДА не хранится в
// открытом виде: дескриптор сразу хеширует введённый пароль перед сохранением.
type Hasher interface {
	Hash(plain string) (string, error)
}

// UserOptions отдаёт варианты для связи «Владелец» в других сущностях (список
// пользователей: id → email). Объявлено здесь, т.к. источник — репозиторий пользователей.
func UserOptions(repo UserAdminRepo) func(ctx context.Context) ([]admin.Option, error) {
	return func(ctx context.Context) ([]admin.Option, error) {
		users, err := repo.ListAll(ctx)
		if err != nil {
			return nil, err
		}
		opts := make([]admin.Option, 0, len(users))
		for _, u := range users {
			opts = append(opts, admin.Option{Value: strconv.FormatInt(u.ID, 10), Label: u.Email})
		}
		return opts, nil
	}
}

// roleOptions — варианты роли пользователя для select.
var roleOptions = []admin.Option{
	{Value: string(domain.RoleUser), Label: "Пользователь"},
	{Value: string(domain.RoleAdmin), Label: "Администратор"},
}

// RegisterUser регистрирует сущность «Пользователи». Особенности против простого
// эталона Note: select-роль, статус-бейдж подтверждения почты в списке, и ПАРОЛЬ —
// обязателен при создании, при редактировании опционален (пустой → не меняем).
func RegisterUser(reg *admin.Registry, repo UserAdminRepo, hasher Hasher) {
	admin.Register(reg, admin.Resource[domain.User]{
		Name:  "users",
		Title: "Пользователи",
		Icon:  "users",
		Fields: []admin.Field{
			{Name: "email", Label: "Email", Type: admin.FieldText, Required: true, InList: true, Sortable: true},
			{Name: "role", Label: "Роль", Type: admin.FieldSelect, Required: true, InList: true, Options: roleOptions},
			{Name: "verified", Label: "Почта подтверждена", Type: admin.FieldBool, InList: true},
			{Name: "password", Label: "Пароль", Type: admin.FieldText, Help: "При редактировании оставьте пустым, чтобы не менять"},
		},

		Row: func(u domain.User) admin.Record {
			roleLabel := "Пользователь"
			tone := ""
			if u.Role == domain.RoleAdmin {
				roleLabel, tone = "Администратор", "info"
			}
			return admin.Record{
				ID: strconv.FormatInt(u.ID, 10),
				Cells: map[string]admin.Cell{
					"email":    {Value: u.Email, Display: u.Email},
					"role":     {Value: string(u.Role), Display: roleLabel, Tone: tone},
					"verified": {Value: strconv.FormatBool(u.EmailVerified())},
				},
			}
		},

		List: func(ctx context.Context) ([]domain.User, error) { return repo.ListAll(ctx) },

		Get: func(ctx context.Context, id string) (admin.FormValues, error) {
			uid, _ := strconv.ParseInt(id, 10, 64)
			u, err := repo.ByID(ctx, uid)
			if err != nil {
				return nil, err
			}
			// Пароль в форму НЕ возвращаем (хэш не показываем) — поле пустое.
			return admin.FormValues{
				"email":    u.Email,
				"role":     string(u.Role),
				"verified": strconv.FormatBool(u.EmailVerified()),
			}, nil
		},

		Create: func(ctx context.Context, v admin.FormValues, _ map[string]admin.FileInput) error {
			// Доменная валидация email/пароля/роли — в одном месте.
			u, err := domain.NewUser(v["email"], v["password"], domain.Role(v["role"]))
			if err != nil {
				return err
			}
			hash, err := hasher.Hash(v["password"])
			if err != nil {
				return err
			}
			u.PasswordHash = hash
			_, err = repo.Create(ctx, u)
			return err
		},

		Update: func(ctx context.Context, id string, v admin.FormValues, _ map[string]admin.FileInput) error {
			uid, _ := strconv.ParseInt(id, 10, 64)
			role := domain.Role(v["role"])
			if !role.Valid() {
				return domain.ErrValidation{Field: "role", Msg: "недопустимая роль"}
			}
			email := domain.NormalizeEmail(v["email"])
			if email == "" {
				return domain.ErrValidation{Field: "email", Msg: "email обязателен"}
			}
			// Email/роль/подтверждение почты — обычным апдейтом. Пароль здесь НЕ меняем
			// (смена пароля — отдельный поток восстановления; в админке оставлено простым).
			verified := v["verified"] == "true"
			return repo.UpdateRoleAndEmail(ctx, uid, email, role, verified)
		},

		Delete: func(ctx context.Context, id string) error {
			uid, _ := strconv.ParseInt(id, 10, 64)
			return repo.DeleteUser(ctx, uid)
		},
	})
}
```

## `internal/domain/email_verification.go`

```go
package domain

import "time"

// EmailVerification — одноразовый токен подтверждения адреса почты с TTL.
// UsedAt непусто → токен уже использован. Отдельная сущность от сброса пароля
// (механизмы независимы), хоть и устроена так же.
type EmailVerification struct {
	Token     string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    time.Time
}

// Usable сообщает, годен ли токен к моменту now: не использован и не истёк.
func (v EmailVerification) Usable(now time.Time) bool {
	return v.UsedAt.IsZero() && now.Before(v.ExpiresAt)
}
```

## `internal/domain/errors.go`

```go
package domain

import (
	"errors"
	"fmt"
	"time"
)

// Доменные ошибки — sentinel-значения и типы, по которым верхние слои принимают
// решения (транспорт мапит их в HTTP-коды). Проверять через errors.Is / errors.As.
//
// ОБРАЗЕЦ: новые предсказуемые ошибки добавляются сюда, чтобы транспортный слой
// мог единообразно их обрабатывать (см. internal/transport/web/errors.go).

// ErrNotFound — запрошенная сущность не существует. → HTTP 404.
type ErrNotFound struct {
	Entity string
	ID     int64
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s с id=%d не найден", e.Entity, e.ID)
}

// ErrValidation — нарушен инвариант сущности (некорректный ввод). → HTTP 400.
type ErrValidation struct {
	Field string
	Msg   string
}

func (e ErrValidation) Error() string {
	if e.Field == "" {
		return e.Msg
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// Ошибки аутентификации/авторизации. Sentinel-значения — проверять через errors.Is.
var (
	// ErrInvalidCredentials — неверный email или пароль. → HTTP 401.
	// Единая ошибка и для «нет такого пользователя», и для «пароль не подошёл» —
	// чтобы по ответу нельзя было определить, существует ли email.
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	// ErrSignupClosed — регистрация выключена настройкой. → HTTP 403.
	ErrSignupClosed = errors.New("регистрация закрыта")
	// ErrEmailTaken — email уже зарегистрирован. → HTTP 409.
	ErrEmailTaken = errors.New("этот email уже зарегистрирован")
	// ErrUnauthenticated — нет валидной сессии. → редирект на вход (или 401 для htmx).
	ErrUnauthenticated = errors.New("требуется вход")
	// ErrForbidden — недостаточно прав (роль). → HTTP 403.
	ErrForbidden = errors.New("недостаточно прав")
	// ErrInvalidToken — токен сброса не найден/просрочен/использован. → HTTP 400.
	ErrInvalidToken = errors.New("ссылка недействительна или устарела")
	// ErrEmailNotVerified — почта не подтверждена, а это требуется настройкой.
	// → вход блокируется, показываем «подтвердите почту».
	ErrEmailNotVerified = errors.New("подтвердите адрес почты по ссылке из письма")
	// ErrSetupClosed — первичная настройка уже завершена (админ создан). → HTTP 410.
	ErrSetupClosed = errors.New("первичная настройка уже завершена")
	// ErrSetupToken — неверный setup-токен. → HTTP 403.
	ErrSetupToken = errors.New("неверный код настройки")
)

// ErrRateLimited — превышен лимит попыток. → HTTP 429. Тип (не sentinel), чтобы нести
// RetryAfter для заголовка Retry-After.
type ErrRateLimited struct {
	RetryAfter time.Duration
}

func (e ErrRateLimited) Error() string {
	return "слишком много попыток, попробуйте позже"
}
```

## `internal/domain/note.go`

```go
// Package domain содержит сущности и бизнес-инварианты приложения.
// Слой не зависит ни от чего, кроме стандартной библиотеки: ни БД, ни HTTP,
// ни сторонних пакетов сюда не протекают. Это ядро чистой архитектуры.
//
// ОБРАЗЕЦ ДЛЯ ГЕНЕРАЦИИ: новая сущность добавляется по аналогии с Note —
// поля, конструктор-валидатор (NewNote), бизнес-правила в методах. Никаких
// json/db-тегов в domain: представления для транспорта и хранилища — отдельные
// структуры в своих слоях, конвертация на стыках.
package domain

import (
	"strings"
	"time"
)

// Note — заметка. Минимальная сущность-образец: её достаточно, чтобы показать
// полный вертикальный срез (domain → usecase → repository → transport).
type Note struct {
	ID        int64
	OwnerID   int64 // владелец (пользователь); проставляет usecase из текущей сессии
	Title     string
	Body      string
	DueDate   string // срок (необязательный), формат YYYY-MM-DD; "" = срока нет
	CreatedAt time.Time
}

// NewNote — конструктор-валидатор. Все инварианты сущности проверяются здесь,
// в одном месте: усечение пробелов, обязательность заголовка, лимит длины.
// ID и CreatedAt проставляются на этапе сохранения (репозиторием/часами).
func NewNote(title, body, dueDate string) (Note, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	dueDate = strings.TrimSpace(dueDate)

	if title == "" {
		return Note{}, ErrValidation{Field: "title", Msg: "заголовок обязателен"}
	}
	if len(title) > 200 {
		return Note{}, ErrValidation{Field: "title", Msg: "заголовок длиннее 200 символов"}
	}
	if len(body) > 10000 {
		return Note{}, ErrValidation{Field: "body", Msg: "текст длиннее 10000 символов"}
	}
	if dueDate != "" {
		if _, err := time.Parse("2006-01-02", dueDate); err != nil {
			return Note{}, ErrValidation{Field: "due_date", Msg: "некорректная дата (нужен формат ГГГГ-ММ-ДД)"}
		}
	}

	return Note{Title: title, Body: body, DueDate: dueDate}, nil
}
```

## `internal/domain/reset.go`

```go
package domain

import "time"

// PasswordReset — одноразовый токен сброса пароля с ограниченным сроком жизни.
// UsedAt непусто → токен уже использован и больше не годен.
type PasswordReset struct {
	Token     string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    time.Time
}

// Usable сообщает, годен ли токен к моменту now: не использован и не истёк.
func (p PasswordReset) Usable(now time.Time) bool {
	return p.UsedAt.IsZero() && now.Before(p.ExpiresAt)
}
```

## `internal/domain/session.go`

```go
package domain

import "time"

// Session — серверная сессия пользователя. Хранится в БД (можно отозвать: логаут,
// бан, смена пароля), токен живёт в httpOnly-cookie.
type Session struct {
	Token     string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Expired сообщает, истекла ли сессия к моменту now.
func (s Session) Expired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}
```

## `internal/domain/setting.go`

```go
// Настройки приложения — типизированный реестр известных ключей. Значения хранятся
// строками, но каждый ключ имеет тип и вид (обычная настройка/секрет). Реестр —
// единственное место, где объявляются доступные настройки: задать неизвестный ключ
// нельзя. Это защищает от опечаток и мусора, и даёт агенту/UI понятный список.
package domain

import (
	"strconv"
	"strings"
	"time"
)

// SettingKind — вид настройки.
type SettingKind string

const (
	// SettingConfig — обычная настройка: значение читается обратно (в API/UI).
	SettingConfig SettingKind = "config"
	// SettingSecret — секрет: значение принимается и используется, но наружу не
	// отдаётся (в ответах — только признак «задано», в логах маскируется).
	SettingSecret SettingKind = "secret"
)

// SettingType — тип значения настройки (для валидации ввода).
type SettingType string

const (
	SettingBool   SettingType = "bool"
	SettingString SettingType = "string"
)

// SettingDef — описание известной настройки в реестре.
type SettingDef struct {
	Key     string
	Kind    SettingKind
	Type    SettingType
	Default string // строковое представление значения по умолчанию
	Title   string // человекочитаемое название (для будущего UI)
}

// settingRegistry — реестр всех доступных настроек приложения. Новая настройка
// добавляется сюда (и больше нигде) — после этого её можно задавать через API.
var settingRegistry = []SettingDef{
	{Key: "allow_signup", Kind: SettingConfig, Type: SettingBool, Default: "true", Title: "Открытая регистрация"},
	{Key: "require_email_verification", Kind: SettingConfig, Type: SettingBool, Default: "false", Title: "Требовать подтверждение почты"},
	{Key: "app_name", Kind: SettingConfig, Type: SettingString, Default: "Zerovibe", Title: "Название приложения"},
}

// SettingDefs возвращает копию реестра (для перечисления в API/UI).
func SettingDefs() []SettingDef {
	out := make([]SettingDef, len(settingRegistry))
	copy(out, settingRegistry)
	return out
}

// LookupSettingDef ищет описание настройки по ключу.
func LookupSettingDef(key string) (SettingDef, bool) {
	for _, d := range settingRegistry {
		if d.Key == key {
			return d, true
		}
	}
	return SettingDef{}, false
}

// Setting — хранимое значение настройки.
type Setting struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

// ValidateSetting проверяет, что ключ известен реестру и значение подходит под тип.
// Возвращает нормализованное значение (например, "true"/"false" для bool) и ошибку
// ErrValidation при нарушении. Хранение/время — забота вызывающего слоя.
func ValidateSetting(key, value string) (string, error) {
	def, ok := LookupSettingDef(key)
	if !ok {
		return "", ErrValidation{Field: "key", Msg: "неизвестная настройка"}
	}
	switch def.Type {
	case SettingBool:
		v := strings.ToLower(strings.TrimSpace(value))
		switch v {
		case "true", "1", "yes", "on":
			return "true", nil
		case "false", "0", "no", "off", "":
			return "false", nil
		default:
			return "", ErrValidation{Field: key, Msg: "ожидается да/нет"}
		}
	case SettingString:
		v := strings.TrimSpace(value)
		if len(v) > 1000 {
			return "", ErrValidation{Field: key, Msg: "значение длиннее 1000 символов"}
		}
		return v, nil
	default:
		return "", ErrValidation{Field: key, Msg: "неподдерживаемый тип настройки"}
	}
}

// BoolValue разбирает строковое значение настройки как bool (для чтения в коде).
func BoolValue(value string) bool {
	b, _ := strconv.ParseBool(strings.TrimSpace(value))
	return b
}
```

## `internal/domain/user.go`

```go
// Пользователь приложения, его роль и инварианты. Слой domain — только stdlib:
// хеширование пароля живёт в usecase (это деталь алгоритма аутентификации), здесь
// валидируется лишь ПЛЕЙН-пароль (границы) и формат email/роли.
package domain

import (
	"strings"
	"time"
)

// Role — роль пользователя. Расширяемая: новая роль добавляется константой и в Valid().
// Состояния доступа в приложении: гость (нет пользователя) → user → admin.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Valid сообщает, известна ли роль. Точка расширения при добавлении новых ролей.
func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleAdmin:
		return true
	default:
		return false
	}
}

// User — учётная запись приложения. PasswordHash проставляет usecase (bcrypt),
// domain его не вычисляет. Без json/db-тегов — представления в своих слоях.
type User struct {
	ID              int64
	Email           string
	PasswordHash    string
	Role            Role
	EmailVerifiedAt time.Time // непусто → почта подтверждена
	CreatedAt       time.Time
}

// EmailVerified сообщает, подтверждена ли почта пользователя.
func (u User) EmailVerified() bool { return !u.EmailVerifiedAt.IsZero() }

// Пароль: минимум — в рунах (дружелюбно к кириллице), максимум — в БАЙТАХ, т.к.
// bcrypt молча игнорирует всё после 72 байт (иначе длинный пароль обрезался бы
// незаметно — это дыра). Поэтому верхнюю границу считаем в байтах.
const (
	minPasswordRunes = 8
	maxPasswordBytes = 72
)

// NewUser — конструктор-валидатор. Нормализует email, проверяет формат, ПЛЕЙН-пароль
// и роль. Хеш и ID/CreatedAt проставляются на следующих слоях (usecase/репозиторий).
func NewUser(email, passwordPlain string, role Role) (User, error) {
	email = NormalizeEmail(email)
	if email == "" {
		return User{}, ErrValidation{Field: "email", Msg: "email обязателен"}
	}
	if !looksLikeEmail(email) {
		return User{}, ErrValidation{Field: "email", Msg: "похоже на некорректный email"}
	}
	if len(email) > 254 {
		return User{}, ErrValidation{Field: "email", Msg: "email слишком длинный"}
	}
	if err := ValidatePasswordPlain(passwordPlain); err != nil {
		return User{}, err
	}
	if !role.Valid() {
		return User{}, ErrValidation{Field: "role", Msg: "недопустимая роль"}
	}
	return User{Email: email, Role: role}, nil
}

// ValidatePasswordPlain проверяет ПЛЕЙН-пароль (длину). Используется и в NewUser, и
// при смене пароля (сброс), где email/роль не трогаются.
func ValidatePasswordPlain(p string) error {
	if len([]rune(p)) < minPasswordRunes {
		return ErrValidation{Field: "password", Msg: "пароль короче 8 символов"}
	}
	if len(p) > maxPasswordBytes {
		return ErrValidation{Field: "password", Msg: "пароль слишком длинный"}
	}
	return nil
}

// NormalizeEmail приводит email к каноничному виду (trim + нижний регистр). Это
// продуктовое упрощение: считаем email кейс-инсенситивным целиком.
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// looksLikeEmail — дешёвая проверка формата без regexp: ровно один '@', непустые
// части до/после, точка в доменной части. Достаточно для отсева опечаток.
func looksLikeEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at != strings.LastIndexByte(s, '@') {
		return false
	}
	local, domain := s[:at], s[at+1:]
	if local == "" || domain == "" {
		return false
	}
	dot := strings.IndexByte(domain, '.')
	return dot > 0 && dot < len(domain)-1
}
```

## `internal/platform/db/db.go`

```go
// Package db — инфраструктурный слой доступа к SQLite с сериализацией записей.
//
// ПОЧЕМУ ОЧЕРЕДЬ ЗАПИСИ. SQLite допускает много одновременных читателей, но
// только одного писателя в момент времени. При параллельных запросах второй
// писатель получает SQLITE_BUSY. Вместо того чтобы ловить эту ошибку и делать
// ретраи в каждом репозитории, мы заводим ОДНУ writer-горутину: все записи
// проходят через неё последовательно. Тогда конкуренции писателей нет в
// принципе, а вызывающий код остаётся простым (просто db.Write(...)).
//
// Это закрывает 99% потребностей небольшого веб-приложения на одном инстансе.
// Чтения идут напрямую (db.Read), параллельно, без очереди.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // чистый Go SQLite-драйвер, без CGO → статический бинарь
)

// DB оборачивает *sql.DB и канал команд записи. Создавать через Open, закрывать
// через Close (он останавливает writer-горутину и закрывает соединение).
type DB struct {
	sql    *sql.DB
	writes chan writeOp
	done   chan struct{}
}

// writeOp — единица работы для writer-горутины: функция записи + канал ответа.
type writeOp struct {
	fn    func(*sql.DB) error
	reply chan error
}

// Open открывает базу по пути dsn (например, "file:app.db") и запускает
// writer-горутину. Включает WAL и busy_timeout как дополнительную страховку.
func Open(dsn string) (*DB, error) {
	// _pragma в DSN драйвера modernc применяются к каждому соединению.
	// Склеиваем корректно: ? если параметров ещё нет, иначе & — чтобы dsn,
	// уже содержащий query-строку, не ломался.
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	full := dsn + sep + "_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)"

	sqlDB, err := sql.Open("sqlite", full)
	if err != nil {
		return nil, fmt.Errorf("открыть sqlite: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	d := &DB{
		sql:    sqlDB,
		writes: make(chan writeOp),
		done:   make(chan struct{}),
	}
	go d.writer()
	return d, nil
}

// writer — единственная горутина, выполняющая записи последовательно.
func (d *DB) writer() {
	for {
		select {
		case op := <-d.writes:
			op.reply <- op.fn(d.sql)
		case <-d.done:
			return
		}
	}
}

// Write ставит операцию записи в очередь и ждёт её завершения. Все INSERT/
// UPDATE/DELETE репозиториев идут через этот метод — так гарантируется один
// писатель. Уважает отмену контекста на время ожидания слота в очереди.
func (d *DB) Write(ctx context.Context, fn func(*sql.DB) error) error {
	op := writeOp{fn: fn, reply: make(chan error, 1)}
	select {
	case d.writes <- op:
	case <-ctx.Done():
		return ctx.Err()
	case <-d.done:
		return fmt.Errorf("db закрыта")
	}
	select {
	case err := <-op.reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Read выполняет чтение напрямую (без очереди) — читатели в SQLite/WAL
// работают параллельно.
func (d *DB) Read(fn func(*sql.DB) error) error {
	return fn(d.sql)
}

// Close останавливает writer-горутину и закрывает соединение.
func (d *DB) Close() error {
	close(d.done)
	return d.sql.Close()
}
```

## `internal/platform/db/migrate.go`

```go
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sync"

	"github.com/pressly/goose/v3"
)

// migrationsFS — встроенные SQL-миграции goose. Лежат рядом, в каталоге migrations/.
// Встраиваем в бинарь, чтобы не зависеть от файловой системы в рантайме (важно для
// distroless-образа без shell). Новая фича = новый файл migrations/NNNNN_*.sql.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// gooseSetup гарантирует одноразовую настройку goose (диалект + источник файлов).
// goose хранит это в глобальном состоянии, поэтому настраиваем под мьютексом и один раз.
var gooseSetup sync.Once

// MigrateUp применяет все непрогнанные миграции (idempotent: уже применённые goose
// пропускает по своей служебной таблице goose_db_version). Вызывается на старте
// приложения до приёма трафика. DDL идёт через очередь записи (db.Write) — единый
// писатель, как и для всех остальных записей.
func (d *DB) MigrateUp(ctx context.Context) error {
	var setupErr error
	gooseSetup.Do(func() {
		goose.SetBaseFS(migrationsFS)
		// modernc.org/sqlite регистрируется как драйвер "sqlite"; goose принимает и
		// "sqlite", и "sqlite3" как один диалект SQLite.
		setupErr = goose.SetDialect("sqlite")
	})
	if setupErr != nil {
		return fmt.Errorf("goose: установить диалект: %w", setupErr)
	}

	return d.Write(ctx, func(s *sql.DB) error {
		if err := goose.UpContext(ctx, s, "migrations"); err != nil {
			return fmt.Errorf("goose: применить миграции: %w", err)
		}
		return nil
	})
}
```

## `internal/platform/db/migrations/00001_init.sql`

```sql
-- Первая миграция: исходная схема приложения (заметки-образец).
-- Совпадает со схемой, которая раньше применялась через CREATE TABLE IF NOT EXISTS,
-- чтобы у уже работающих приложений база не разошлась при переходе на goose.
-- +goose Up
CREATE TABLE IF NOT EXISTS notes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS notes;
```

## `internal/platform/db/migrations/00002_settings.sql`

```sql
-- Настройки приложения: key-value с типизацией в коде (реестр доменных настроек).
-- +goose Up
CREATE TABLE IF NOT EXISTS settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS settings;
```

## `internal/platform/db/migrations/00003_auth.sql`

```sql
-- Аутентификация: пользователи, сессии, токены сброса пароля, счётчики рейт-лимита.
-- Плюс заметки становятся личными (привязка к владельцу).
-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user',
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

CREATE TABLE IF NOT EXISTS password_resets (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL,
    used_at    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_resets_user ON password_resets(user_id);

CREATE TABLE IF NOT EXISTS rate_counters (
    key          TEXT PRIMARY KEY,
    window_start TEXT NOT NULL,
    count        INTEGER NOT NULL DEFAULT 0
);

-- Заметки получают владельца. owner_id=0 — «ничей» (для возможных старых записей до
-- появления аутентификации); реальные пользователи начинаются с id=1.
ALTER TABLE notes ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_notes_owner ON notes(owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_notes_owner;
ALTER TABLE notes DROP COLUMN owner_id;
DROP TABLE IF EXISTS rate_counters;
DROP INDEX IF EXISTS idx_resets_user;
DROP TABLE IF EXISTS password_resets;
DROP INDEX IF EXISTS idx_sessions_user;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
```

## `internal/platform/db/migrations/00004_email_verification.sql`

```sql
-- Подтверждение почты: токены подтверждения + отметка у пользователя.
-- +goose Up
CREATE TABLE IF NOT EXISTS email_verifications (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL,
    used_at    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_email_verifications_user ON email_verifications(user_id);

-- Отметка времени подтверждения почты у пользователя (пусто = не подтверждена).
ALTER TABLE users ADD COLUMN email_verified_at TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN email_verified_at;
DROP INDEX IF EXISTS idx_email_verifications_user;
DROP TABLE IF EXISTS email_verifications;
```

## `internal/platform/db/migrations/00005_note_due_date.sql`

```sql
-- +goose Up
-- Срок заметки (необязательная дата) — демонстрирует редактируемое date-поле в админке
-- (нативный календарь). Пусто = срока нет.
ALTER TABLE notes ADD COLUMN due_date TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE notes DROP COLUMN due_date;
```

## `internal/repository/sqlite/email_verifications.go`

```go
// SQLite-репозиторий токенов подтверждения почты. used_at=” = не использован.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/chudno/zerovibe/internal/domain"
)

// EmailVerificationRepo — SQLite-репозиторий токенов подтверждения почты.
type EmailVerificationRepo struct {
	db writer
}

// NewEmailVerificationRepo собирает репозиторий поверх платформенного db.
func NewEmailVerificationRepo(db writer) *EmailVerificationRepo {
	return &EmailVerificationRepo{db: db}
}

// Create сохраняет токен подтверждения.
func (r *EmailVerificationRepo) Create(ctx context.Context, v domain.EmailVerification) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx,
			`INSERT INTO email_verifications (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
			v.Token, v.UserID, v.CreatedAt.UTC().Format(sqliteTime), v.ExpiresAt.UTC().Format(sqliteTime),
		)
		if err != nil {
			return fmt.Errorf("insert email verification: %w", err)
		}
		return nil
	})
}

// ByToken находит токен подтверждения; ErrNotFound если нет.
func (r *EmailVerificationRepo) ByToken(ctx context.Context, token string) (domain.EmailVerification, error) {
	var v domain.EmailVerification
	err := r.db.Read(func(s *sql.DB) error {
		var created, expires, used string
		err := s.QueryRowContext(ctx,
			`SELECT token, user_id, created_at, expires_at, used_at FROM email_verifications WHERE token = ?`, token,
		).Scan(&v.Token, &v.UserID, &created, &expires, &used)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound{Entity: "email verification"}
		}
		if err != nil {
			return fmt.Errorf("select email verification: %w", err)
		}
		v.CreatedAt = parseTime(created)
		v.ExpiresAt = parseTime(expires)
		if used != "" {
			v.UsedAt = parseTime(used)
		}
		return nil
	})
	if err != nil {
		return domain.EmailVerification{}, err
	}
	return v, nil
}

// MarkUsed помечает токен использованным.
func (r *EmailVerificationRepo) MarkUsed(ctx context.Context, token string, usedAt time.Time) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx,
			`UPDATE email_verifications SET used_at = ? WHERE token = ?`,
			usedAt.UTC().Format(sqliteTime), token,
		)
		if err != nil {
			return fmt.Errorf("mark email verification used: %w", err)
		}
		return nil
	})
}

// DeleteByUser удаляет все токены подтверждения пользователя (выдача нового
// инвалидирует старые).
func (r *EmailVerificationRepo) DeleteByUser(ctx context.Context, userID int64) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx, `DELETE FROM email_verifications WHERE user_id = ?`, userID)
		if err != nil {
			return fmt.Errorf("delete email verifications by user: %w", err)
		}
		return nil
	})
}
```

## `internal/repository/sqlite/notes.go`

```go
// Package sqlite — реализация портов usecase поверх SQLite.
//
// КЛЮЧЕВОЙ ПАТТЕРН: записи (Create/Delete) идут через db.Write — попадают в
// единую writer-горутину и сериализуются (нет SQLITE_BUSY). Чтения (List) идут
// через db.Read — параллельно. Конвертация domain↔строка БД — в этом слое.
//
// ОБРАЗЕЦ ДЛЯ ГЕНЕРАЦИИ: на каждую сущность — свой репозиторий, реализующий
// порт из usecase. INSERT/UPDATE/DELETE оборачивать в db.Write, SELECT — в db.Read.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chudno/zerovibe/internal/domain"
)

// writer — минимальный интерфейс к платформенному db.DB (Write/Read).
// Принимаем интерфейсом, а не *db.DB, чтобы репозиторий было легко тестировать.
type writer interface {
	Write(ctx context.Context, fn func(*sql.DB) error) error
	Read(fn func(*sql.DB) error) error
}

// NoteRepo — SQLite-репозиторий заметок.
type NoteRepo struct {
	db writer
}

// NewNoteRepo собирает репозиторий поверх платформенного db.
func NewNoteRepo(db writer) *NoteRepo {
	return &NoteRepo{db: db}
}

// Create вставляет заметку владельца (через очередь записи) и возвращает её с id и
// проставленным базой created_at — чтобы фрагмент сразу показывал время.
func (r *NoteRepo) Create(ctx context.Context, n domain.Note) (domain.Note, error) {
	err := r.db.Write(ctx, func(s *sql.DB) error {
		var created string
		// RETURNING поддерживается SQLite ≥3.35 (есть в modernc) — отдаёт id и
		// время одним запросом, без отдельного SELECT.
		err := s.QueryRowContext(ctx,
			`INSERT INTO notes (owner_id, title, body, due_date) VALUES (?, ?, ?, ?) RETURNING id, created_at`,
			n.OwnerID, n.Title, n.Body, n.DueDate,
		).Scan(&n.ID, &created)
		if err != nil {
			return fmt.Errorf("insert note: %w", err)
		}
		n.CreatedAt = parseTime(created)
		return nil
	})
	if err != nil {
		return domain.Note{}, err
	}
	return n, nil
}

// ListByOwner возвращает заметки владельца, новые сверху.
func (r *NoteRepo) ListByOwner(ctx context.Context, ownerID int64) ([]domain.Note, error) {
	var notes []domain.Note
	err := r.db.Read(func(s *sql.DB) error {
		rows, err := s.QueryContext(ctx,
			`SELECT id, owner_id, title, body, due_date, created_at FROM notes WHERE owner_id = ? ORDER BY id DESC`, ownerID)
		if err != nil {
			return fmt.Errorf("select notes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var n domain.Note
			var created string
			if err := rows.Scan(&n.ID, &n.OwnerID, &n.Title, &n.Body, &n.DueDate, &created); err != nil {
				return fmt.Errorf("scan note: %w", err)
			}
			n.CreatedAt = parseTime(created)
			notes = append(notes, n)
		}
		return rows.Err()
	})
	return notes, err
}

// Delete удаляет заметку владельца по id (чужую не трогает — отдаёт ErrNotFound,
// скрывая существование чужих заметок).
func (r *NoteRepo) Delete(ctx context.Context, id, ownerID int64) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		res, err := s.ExecContext(ctx, `DELETE FROM notes WHERE id = ? AND owner_id = ?`, id, ownerID)
		if err != nil {
			return fmt.Errorf("delete note: %w", err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return domain.ErrNotFound{Entity: "note", ID: id}
		}
		return nil
	})
}
```

## `internal/repository/sqlite/notes_admin.go`

```go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chudno/zerovibe/internal/domain"
)

// Admin-методы NoteRepo: доступ ко ВСЕМ заметкам без фильтра по владельцу. Их
// использует встроенная админка (internal/admin) — администратор видит и правит
// данные всех пользователей. Владельческие методы (ListByOwner/Delete) остаются
// для обычного пользовательского потока. ОБРАЗЕЦ: новая сущность под админку
// получает такой же набор ListAll/GetByID/UpdateAny/DeleteAny.

// ListAll возвращает все заметки (новые сверху) — для списка в админке.
func (r *NoteRepo) ListAll(ctx context.Context) ([]domain.Note, error) {
	var notes []domain.Note
	err := r.db.Read(func(s *sql.DB) error {
		rows, err := s.QueryContext(ctx,
			`SELECT id, owner_id, title, body, due_date, created_at FROM notes ORDER BY id DESC`)
		if err != nil {
			return fmt.Errorf("select all notes: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var n domain.Note
			var created string
			if err := rows.Scan(&n.ID, &n.OwnerID, &n.Title, &n.Body, &n.DueDate, &created); err != nil {
				return fmt.Errorf("scan note: %w", err)
			}
			n.CreatedAt = parseTime(created)
			notes = append(notes, n)
		}
		return rows.Err()
	})
	return notes, err
}

// GetByID возвращает заметку по id (для формы редактирования в админке).
func (r *NoteRepo) GetByID(ctx context.Context, id int64) (domain.Note, error) {
	var n domain.Note
	err := r.db.Read(func(s *sql.DB) error {
		var created string
		err := s.QueryRowContext(ctx,
			`SELECT id, owner_id, title, body, due_date, created_at FROM notes WHERE id = ?`, id).
			Scan(&n.ID, &n.OwnerID, &n.Title, &n.Body, &n.DueDate, &created)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound{Entity: "note", ID: id}
		}
		if err != nil {
			return fmt.Errorf("get note: %w", err)
		}
		n.CreatedAt = parseTime(created)
		return nil
	})
	return n, err
}

// UpdateAny обновляет заголовок/текст заметки по id (админ правит любую).
func (r *NoteRepo) UpdateAny(ctx context.Context, n domain.Note) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		res, err := s.ExecContext(ctx,
			`UPDATE notes SET title = ?, body = ?, due_date = ? WHERE id = ?`, n.Title, n.Body, n.DueDate, n.ID)
		if err != nil {
			return fmt.Errorf("update note: %w", err)
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			return domain.ErrNotFound{Entity: "note", ID: n.ID}
		}
		return nil
	})
}

// DeleteAny удаляет заметку по id без проверки владельца (админ удаляет любую).
func (r *NoteRepo) DeleteAny(ctx context.Context, id int64) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		res, err := s.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("delete note: %w", err)
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			return domain.ErrNotFound{Entity: "note", ID: id}
		}
		return nil
	})
}
```

## `internal/repository/sqlite/ratelimit.go`

```go
// SQLite-репозиторий рейт-лимитов (оконные счётчики). Учёт попытки — атомарный
// read-modify-write СТРОГО внутри одного db.Write: единственная writer-горутина
// сериализует записи, поэтому между чтением и обновлением счётчика не вклинится
// другой писатель — гонок нет без явных транзакций. Счётчики переживают рестарт
// (хранятся в БД), как и требовалось.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// RateLimitRepo — SQLite-репозиторий оконных счётчиков.
type RateLimitRepo struct {
	db writer
}

// NewRateLimitRepo собирает репозиторий рейт-лимитов поверх платформенного db.
func NewRateLimitRepo(db writer) *RateLimitRepo {
	return &RateLimitRepo{db: db}
}

// Allow учитывает попытку по ключу и сообщает, не превышен ли лимit. Fixed-window:
// если текущее окно истекло (now - window_start >= window) — начинаем новое.
// Возвращает (разрешено, сколько ждать до сброса окна при запрете, ошибка).
func (r *RateLimitRepo) Allow(ctx context.Context, key string, limit int, window time.Duration, now time.Time) (bool, time.Duration, error) {
	var allowed bool
	var retryAfter time.Duration

	err := r.db.Write(ctx, func(s *sql.DB) error {
		var startStr string
		var count int
		err := s.QueryRowContext(ctx,
			`SELECT window_start, count FROM rate_counters WHERE key = ?`, key,
		).Scan(&startStr, &count)

		windowStart := now
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// первой попытки ещё не было — заводим окно
			count = 0
			windowStart = now
		case err != nil:
			return fmt.Errorf("select rate counter: %w", err)
		default:
			windowStart = parseTime(startStr)
			if now.Sub(windowStart) >= window {
				// окно истекло — новое окно
				count = 0
				windowStart = now
			}
		}

		count++
		allowed = count <= limit
		if !allowed {
			retryAfter = window - now.Sub(windowStart)
			if retryAfter < 0 {
				retryAfter = 0
			}
		}

		_, err = s.ExecContext(ctx,
			`INSERT INTO rate_counters (key, window_start, count) VALUES (?, ?, ?)
			 ON CONFLICT(key) DO UPDATE SET window_start = excluded.window_start, count = excluded.count`,
			key, windowStart.UTC().Format(sqliteTime), count,
		)
		if err != nil {
			return fmt.Errorf("upsert rate counter: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	return allowed, retryAfter, nil
}
```

## `internal/repository/sqlite/resets.go`

```go
// SQLite-репозиторий токенов сброса пароля. used_at=” означает «не использован».
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/chudno/zerovibe/internal/domain"
)

// ResetRepo — SQLite-репозиторий токенов сброса.
type ResetRepo struct {
	db writer
}

// NewResetRepo собирает репозиторий токенов сброса поверх платформенного db.
func NewResetRepo(db writer) *ResetRepo {
	return &ResetRepo{db: db}
}

// Create сохраняет токен сброса.
func (r *ResetRepo) Create(ctx context.Context, p domain.PasswordReset) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx,
			`INSERT INTO password_resets (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
			p.Token, p.UserID, p.CreatedAt.UTC().Format(sqliteTime), p.ExpiresAt.UTC().Format(sqliteTime),
		)
		if err != nil {
			return fmt.Errorf("insert reset: %w", err)
		}
		return nil
	})
}

// ByToken находит токен сброса; ErrNotFound если нет.
func (r *ResetRepo) ByToken(ctx context.Context, token string) (domain.PasswordReset, error) {
	var p domain.PasswordReset
	err := r.db.Read(func(s *sql.DB) error {
		var created, expires, used string
		err := s.QueryRowContext(ctx,
			`SELECT token, user_id, created_at, expires_at, used_at FROM password_resets WHERE token = ?`, token,
		).Scan(&p.Token, &p.UserID, &created, &expires, &used)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound{Entity: "reset"}
		}
		if err != nil {
			return fmt.Errorf("select reset: %w", err)
		}
		p.CreatedAt = parseTime(created)
		p.ExpiresAt = parseTime(expires)
		if used != "" {
			p.UsedAt = parseTime(used)
		}
		return nil
	})
	if err != nil {
		return domain.PasswordReset{}, err
	}
	return p, nil
}

// MarkUsed помечает токен использованным.
func (r *ResetRepo) MarkUsed(ctx context.Context, token string, usedAt time.Time) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx,
			`UPDATE password_resets SET used_at = ? WHERE token = ?`,
			usedAt.UTC().Format(sqliteTime), token,
		)
		if err != nil {
			return fmt.Errorf("mark reset used: %w", err)
		}
		return nil
	})
}

// DeleteByUser удаляет все токены сброса пользователя (выдача нового инвалидирует старые).
func (r *ResetRepo) DeleteByUser(ctx context.Context, userID int64) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx, `DELETE FROM password_resets WHERE user_id = ?`, userID)
		if err != nil {
			return fmt.Errorf("delete resets by user: %w", err)
		}
		return nil
	})
}
```

## `internal/repository/sqlite/sessions.go`

```go
// SQLite-репозиторий сессий. Время хранится в UTC в формате datetime('now'),
// чтобы запись/чтение/сравнение были консистентны.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/chudno/zerovibe/internal/domain"
)

// SessionRepo — SQLite-репозиторий сессий.
type SessionRepo struct {
	db writer
}

// NewSessionRepo собирает репозиторий сессий поверх платформенного db.
func NewSessionRepo(db writer) *SessionRepo {
	return &SessionRepo{db: db}
}

// Create сохраняет сессию.
func (r *SessionRepo) Create(ctx context.Context, sess domain.Session) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx,
			`INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
			sess.Token, sess.UserID,
			sess.CreatedAt.UTC().Format(sqliteTime), sess.ExpiresAt.UTC().Format(sqliteTime),
		)
		if err != nil {
			return fmt.Errorf("insert session: %w", err)
		}
		return nil
	})
}

// ByToken находит сессию по токену; ErrNotFound если нет.
func (r *SessionRepo) ByToken(ctx context.Context, token string) (domain.Session, error) {
	var sess domain.Session
	err := r.db.Read(func(s *sql.DB) error {
		var created, expires string
		err := s.QueryRowContext(ctx,
			`SELECT token, user_id, created_at, expires_at FROM sessions WHERE token = ?`, token,
		).Scan(&sess.Token, &sess.UserID, &created, &expires)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound{Entity: "session"}
		}
		if err != nil {
			return fmt.Errorf("select session: %w", err)
		}
		sess.CreatedAt = parseTime(created)
		sess.ExpiresAt = parseTime(expires)
		return nil
	})
	if err != nil {
		return domain.Session{}, err
	}
	return sess, nil
}

// Delete удаляет сессию по токену (идемпотентно на уровне SQL — отсутствие не ошибка).
func (r *SessionRepo) Delete(ctx context.Context, token string) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
		if err != nil {
			return fmt.Errorf("delete session: %w", err)
		}
		return nil
	})
}

// DeleteByUser удаляет все сессии пользователя (бан/смена пароля/«выйти везде»).
func (r *SessionRepo) DeleteByUser(ctx context.Context, userID int64) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
		if err != nil {
			return fmt.Errorf("delete sessions by user: %w", err)
		}
		return nil
	})
}

// DeleteExpired удаляет истёкшие сессии (фоновый GC).
func (r *SessionRepo) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, now.UTC().Format(sqliteTime))
		if err != nil {
			return fmt.Errorf("delete expired sessions: %w", err)
		}
		return nil
	})
}
```

## `internal/repository/sqlite/settings.go`

```go
// SQLite-репозиторий настроек приложения. Запись (Set) — через очередь записи
// (db.Write), чтения (Get/List) — через db.Read. Схема таблицы settings заводится
// goose-миграцией, здесь только доступ.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chudno/zerovibe/internal/domain"
)

// SettingRepo — SQLite-репозиторий настроек.
type SettingRepo struct {
	db writer
}

// NewSettingRepo собирает репозиторий настроек поверх платформенного db.
func NewSettingRepo(db writer) *SettingRepo {
	return &SettingRepo{db: db}
}

// Get возвращает настройку по ключу; ErrNotFound, если не задана.
func (r *SettingRepo) Get(ctx context.Context, key string) (domain.Setting, error) {
	var st domain.Setting
	err := r.db.Read(func(s *sql.DB) error {
		var updated string
		err := s.QueryRowContext(ctx,
			`SELECT key, value, updated_at FROM settings WHERE key = ?`, key,
		).Scan(&st.Key, &st.Value, &updated)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound{Entity: "setting"}
		}
		if err != nil {
			return fmt.Errorf("select setting: %w", err)
		}
		st.UpdatedAt = parseTime(updated)
		return nil
	})
	if err != nil {
		return domain.Setting{}, err
	}
	return st, nil
}

// Set вставляет или обновляет настройку (UPSERT по ключу).
func (r *SettingRepo) Set(ctx context.Context, st domain.Setting) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		_, err := s.ExecContext(ctx,
			`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
			st.Key, st.Value, st.UpdatedAt.UTC().Format(sqliteTime),
		)
		if err != nil {
			return fmt.Errorf("upsert setting: %w", err)
		}
		return nil
	})
}

// List возвращает все заданные настройки.
func (r *SettingRepo) List(ctx context.Context) ([]domain.Setting, error) {
	var out []domain.Setting
	err := r.db.Read(func(s *sql.DB) error {
		rows, err := s.QueryContext(ctx, `SELECT key, value, updated_at FROM settings`)
		if err != nil {
			return fmt.Errorf("select settings: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var st domain.Setting
			var updated string
			if err := rows.Scan(&st.Key, &st.Value, &updated); err != nil {
				return fmt.Errorf("scan setting: %w", err)
			}
			st.UpdatedAt = parseTime(updated)
			out = append(out, st)
		}
		return rows.Err()
	})
	return out, err
}
```

## `internal/repository/sqlite/sqlite.go`

```go
package sqlite

import (
	"log/slog"
	"time"
)

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
```

## `internal/repository/sqlite/users.go`

```go
// SQLite-репозиторий пользователей. Записи — через db.Write, чтения — через db.Read.
// Конфликт уникальности email мапится в domain.ErrEmailTaken.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chudno/zerovibe/internal/domain"
)

// UserRepo — SQLite-репозиторий пользователей.
type UserRepo struct {
	db writer
}

// NewUserRepo собирает репозиторий пользователей поверх платформенного db.
func NewUserRepo(db writer) *UserRepo {
	return &UserRepo{db: db}
}

// Create вставляет пользователя; дубль email → domain.ErrEmailTaken.
func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	err := r.db.Write(ctx, func(s *sql.DB) error {
		var created string
		err := s.QueryRowContext(ctx,
			`INSERT INTO users (email, password_hash, role) VALUES (?, ?, ?)
			 RETURNING id, created_at`,
			u.Email, u.PasswordHash, string(u.Role),
		).Scan(&u.ID, &created)
		if err != nil {
			if isUniqueViolation(err) {
				return domain.ErrEmailTaken
			}
			return fmt.Errorf("insert user: %w", err)
		}
		u.CreatedAt = parseTime(created)
		return nil
	})
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

// ByEmail находит пользователя по email; ErrNotFound если нет.
func (r *UserRepo) ByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.scanOne(ctx, `SELECT id, email, password_hash, role, email_verified_at, created_at FROM users WHERE email = ?`, email)
}

// ByID находит пользователя по id; ErrNotFound если нет.
func (r *UserRepo) ByID(ctx context.Context, id int64) (domain.User, error) {
	return r.scanOne(ctx, `SELECT id, email, password_hash, role, email_verified_at, created_at FROM users WHERE id = ?`, id)
}

// scanOne читает одного пользователя по запросу с одним аргументом.
func (r *UserRepo) scanOne(ctx context.Context, query string, arg any) (domain.User, error) {
	var u domain.User
	err := r.db.Read(func(s *sql.DB) error {
		var role, verified, created string
		err := s.QueryRowContext(ctx, query, arg).
			Scan(&u.ID, &u.Email, &u.PasswordHash, &role, &verified, &created)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound{Entity: "user"}
		}
		if err != nil {
			return fmt.Errorf("select user: %w", err)
		}
		u.Role = domain.Role(role)
		if verified != "" {
			u.EmailVerifiedAt = parseTime(verified)
		}
		u.CreatedAt = parseTime(created)
		return nil
	})
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

// UpdatePasswordHash меняет хеш пароля; ErrNotFound если пользователя нет.
func (r *UserRepo) UpdatePasswordHash(ctx context.Context, userID int64, hash string) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		res, err := s.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, userID)
		if err != nil {
			return fmt.Errorf("update password: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return domain.ErrNotFound{Entity: "user", ID: userID}
		}
		return nil
	})
}

// MarkEmailVerified проставляет время подтверждения почты; ErrNotFound если нет.
func (r *UserRepo) MarkEmailVerified(ctx context.Context, userID int64, at time.Time) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		res, err := s.ExecContext(ctx,
			`UPDATE users SET email_verified_at = ? WHERE id = ?`,
			at.UTC().Format(sqliteTime), userID)
		if err != nil {
			return fmt.Errorf("mark email verified: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return domain.ErrNotFound{Entity: "user", ID: userID}
		}
		return nil
	})
}

// CountAdmins возвращает число пользователей с ролью admin (для сида первого админа).
func (r *UserRepo) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := r.db.Read(func(s *sql.DB) error {
		return s.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = ?`, string(domain.RoleAdmin)).Scan(&n)
	})
	return n, err
}

// isUniqueViolation распознаёт нарушение UNIQUE-ограничения SQLite (modernc отдаёт
// сообщение вида "UNIQUE constraint failed: users.email").
func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}
```

## `internal/repository/sqlite/users_admin.go`

```go
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chudno/zerovibe/internal/domain"
)

// Admin-методы UserRepo: для встроенной админки (управление пользователями).
// ОСТОРОЖНО с аутентификацией: пароль обновляется ОТДЕЛЬНЫМ методом (а не через
// общий UPDATE), чтобы случайно не затереть хэш; удаление защищает последнего админа.

// ListAll возвращает всех пользователей (новые сверху) — для списка в админке.
func (r *UserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Read(func(s *sql.DB) error {
		rows, err := s.QueryContext(ctx,
			`SELECT id, email, password_hash, role, email_verified_at, created_at FROM users ORDER BY id DESC`)
		if err != nil {
			return fmt.Errorf("select all users: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			u, err := scanUserRow(rows)
			if err != nil {
				return err
			}
			users = append(users, u)
		}
		return rows.Err()
	})
	return users, err
}

// UpdateRoleAndEmail обновляет email, роль и статус подтверждения почты (НЕ трогает
// пароль). verified=true → проставляет email_verified_at (сохраняя уже имеющуюся дату,
// если почта была подтверждена раньше); verified=false → снимает подтверждение.
func (r *UserRepo) UpdateRoleAndEmail(ctx context.Context, id int64, email string, role domain.Role, verified bool) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		now := time.Now().UTC().Format(sqliteTime)
		// CASE сохраняет исходную дату подтверждения, если она уже была (не перетираем
		// её текущим временем при простом сохранении формы).
		res, err := s.ExecContext(ctx,
			`UPDATE users SET email = ?, role = ?,
			   email_verified_at = CASE
			     WHEN ? = 0 THEN ''
			     WHEN email_verified_at = '' THEN ?
			     ELSE email_verified_at
			   END
			 WHERE id = ?`,
			email, string(role), boolToInt(verified), now, id)
		if err != nil {
			if isUniqueViolation(err) {
				return domain.ErrEmailTaken
			}
			return fmt.Errorf("update user: %w", err)
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			return domain.ErrNotFound{Entity: "user", ID: id}
		}
		return nil
	})
}

// boolToInt — 1/0 для SQL-CASE (SQLite не имеет булева типа в параметрах).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// DeleteUser удаляет пользователя по id, НЕ позволяя удалить последнего администратора
// (иначе в приложение будет не войти под админом). Проверка и удаление — в одной
// записи (writer-горутина сериализована, гонки нет).
func (r *UserRepo) DeleteUser(ctx context.Context, id int64) error {
	return r.db.Write(ctx, func(s *sql.DB) error {
		var role string
		if err := s.QueryRowContext(ctx, `SELECT role FROM users WHERE id = ?`, id).Scan(&role); err != nil {
			if err == sql.ErrNoRows {
				return domain.ErrNotFound{Entity: "user", ID: id}
			}
			return fmt.Errorf("get user role: %w", err)
		}
		if domain.Role(role) == domain.RoleAdmin {
			var admins int
			if err := s.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&admins); err != nil {
				return fmt.Errorf("count admins: %w", err)
			}
			if admins <= 1 {
				return domain.ErrValidation{Field: "role", Msg: "нельзя удалить последнего администратора"}
			}
		}
		if _, err := s.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id); err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	})
}

// scanUserRow читает пользователя из *sql.Rows (порядок колонок как в ListAll).
func scanUserRow(rows *sql.Rows) (domain.User, error) {
	var u domain.User
	var role string
	var verified sql.NullString
	var created string
	if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &role, &verified, &created); err != nil {
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	u.Role = domain.Role(role)
	if verified.Valid {
		u.EmailVerifiedAt = parseTime(verified.String)
	}
	u.CreatedAt = parseTime(created)
	return u, nil
}
```

## `internal/transport/web/templates/forgot.html`

```html
{{define "forgot"}}
<div class="mx-auto mt-6 w-full max-w-sm">
	<div class="card border border-base-300 bg-base-100">
		<div class="card-body gap-4">
			{{if .Flash}}
			<div class="alert alert-info text-sm">{{.Flash}}</div>
			<div class="text-sm"><a class="link link-hover" href="/login">Вернуться ко входу</a></div>
			{{else}}
			{{if .Err}}<div class="alert alert-error text-sm">{{.Err}}</div>{{end}}
			<p class="text-sm text-base-content/70">Укажите email — пришлём ссылку для сброса пароля.</p>
			<form method="post" action="/forgot">
				<fieldset class="fieldset">
					<label class="label" for="email">Email</label>
					<input id="email" class="input w-full" type="email" name="email" autocomplete="username" required autofocus>
					<button class="btn btn-primary btn-block mt-2" type="submit">Отправить ссылку</button>
				</fieldset>
			</form>
			<div class="text-sm"><a class="link link-hover" href="/login">Вспомнили пароль? Войти</a></div>
			{{end}}
		</div>
	</div>
</div>
{{end}}
```

## `internal/transport/web/templates/landing.html`

```html
{{define "landing"}}
<div class="flex min-h-screen flex-col bg-base-100 text-base-content">
	{{template "site-header" .}}

	<main class="flex-1">
		<!-- Hero: продаёт САМ шаблон. Техника — словами выгоды (не «Go/HTMX», а
		     «быстро, надёжно, готово из коробки»). Один акцент (primary на главное
		     действие), крупный заголовок с воздухом, выравнивание по левому краю. -->
		<section class="mx-auto w-full max-w-6xl px-6 py-20 md:py-28 lg:px-8">
			<div class="grid items-center gap-12 lg:grid-cols-2">
				<div class="max-w-xl">
					<h1 class="font-display text-4xl font-bold leading-tight tracking-tight text-balance sm:text-5xl lg:text-6xl">
						Всё, что нужно приложению, — уже здесь
					</h1>
					<p class="mt-6 text-lg text-base-content/70">
						Вход, регистрация, личные кабинеты, панель управления и опрятный
						интерфейс работают из коробки. Не нужно собирать это заново — начните
						сразу с того, что делает ваш продукт особенным.
					</p>
					<div class="mt-10 flex flex-wrap items-center gap-4">
						{{if .User}}
						<a class="btn btn-primary btn-lg" href="/">Открыть приложение</a>
						{{else}}
						<a class="btn btn-primary btn-lg" href="/register">Попробовать бесплатно</a>
						<a class="btn btn-ghost btn-lg" href="/login">Войти</a>
						{{end}}
					</div>
				</div>

<!-- Сигнатурный визуал: живая «панель данных» приложения — намёк на бэкофис
				     из коробки (аккаунты, таблица записей, мгновенный отклик, темы), а не
				     стоковая иллюстрация. Анимация — чистый CSS, с prefers-reduced-motion. -->
				<div class="relative">
					<style>
						@keyframes zvPanelIn {
							from { opacity: 0; transform: translateY(18px); }
							to   { opacity: 1; transform: translateY(0); }
						}
						@keyframes zvRise {
							from { opacity: 0; transform: translateY(10px); }
							to   { opacity: 1; transform: translateY(0); }
						}
						@keyframes zvRowGlow {
							0%   { background-color: transparent; }
							12%  { background-color: color-mix(in oklch, var(--color-primary) 12%, transparent); }
							55%  { background-color: color-mix(in oklch, var(--color-primary) 12%, transparent); }
							100% { background-color: transparent; }
						}
						@keyframes zvRowMark {
							0%, 62% { opacity: 1; transform: scale(1); }
							100%    { opacity: 0; transform: scale(0.4); }
						}
						@keyframes zvPulse {
							0%, 100% { opacity: 1; transform: scale(1); }
							50%      { opacity: 0.4; transform: scale(0.7); }
						}
						@keyframes zvRing {
							0%        { opacity: 0.5; transform: scale(1); }
							70%, 100% { opacity: 0; transform: scale(2.6); }
						}
						.zv-panel   { animation: zvPanelIn 0.7s cubic-bezier(0.22, 1, 0.36, 1) both; }
						.zv-rise    { animation: zvRise 0.6s cubic-bezier(0.22, 1, 0.36, 1) both; }
						.zv-d1 { animation-delay: 0.30s; }
						.zv-d2 { animation-delay: 0.40s; }
						.zv-d3 { animation-delay: 0.50s; }
						.zv-d4 { animation-delay: 0.60s; }
						.zv-row-live { animation: zvRowGlow 4.6s ease-in-out infinite; }
						.zv-row-mark { animation: zvRowMark 4.6s ease-in-out infinite; }
						.zv-dot  { animation: zvPulse 2.2s ease-in-out infinite; }
						.zv-ring { animation: zvRing 2.2s ease-out infinite; }

						@media (prefers-reduced-motion: reduce) {
							.zv-panel, .zv-rise, .zv-row-live, .zv-row-mark, .zv-dot, .zv-ring {
								animation: none !important;
							}
							.zv-rise { opacity: 1; transform: none; }
						}
					</style>

					<!-- мягкое акцентное свечение за панелью -->
					<div aria-hidden="true"
					     class="pointer-events-none absolute -inset-6 -z-10 rounded-[2rem] bg-primary/10 blur-3xl"></div>

					<div class="zv-panel overflow-hidden rounded-2xl border border-base-300 bg-base-100 shadow-xl shadow-base-300/40">

						<!-- окно: кружки + адрес + кластер аккаунта -->
						<div class="flex items-center gap-3 border-b border-base-300 bg-base-200/60 px-4 py-3">
							<div class="flex shrink-0 items-center gap-1.5">
								<span class="size-2.5 rounded-full bg-base-300"></span>
								<span class="size-2.5 rounded-full bg-base-300"></span>
								<span class="size-2.5 rounded-full bg-base-300"></span>
							</div>

							<div class="ml-1 flex min-w-0 flex-1 items-center gap-1.5 rounded-lg border border-base-300 bg-base-100 px-2.5 py-1">
								<svg xmlns="http://www.w3.org/2000/svg" class="size-3.5 shrink-0 text-base-content/40" viewBox="0 0 24 24"
								     fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<rect width="18" height="11" x="3" y="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
								</svg>
<span class="truncate font-mono text-xs text-base-content/50">project.zerovibe.app</span>
							</div>

							<!-- кластер аккаунта: пульс «онлайн» + аватар (аккаунты из коробки) -->
							<div class="flex shrink-0 items-center gap-2">
								<span class="relative flex size-2">
									<span class="zv-ring absolute inline-flex size-2 rounded-full bg-primary"></span>
									<span class="zv-dot relative inline-flex size-2 rounded-full bg-primary"></span>
								</span>
								<span class="grid size-6 place-items-center rounded-full bg-primary/15 text-[0.65rem] font-semibold text-primary">АК</span>
							</div>
						</div>

<!-- шапка раздела: заголовок + чип выгоды + счётчик записей -->
						<div class="flex items-center justify-between gap-3 px-5 pt-4 pb-2">
							<div class="flex min-w-0 items-center gap-2">
								<h3 class="text-sm font-semibold text-base-content">Клиенты</h3>
								<span class="hidden items-center gap-1 rounded-full border border-base-300 bg-base-200/70 px-2 py-0.5 text-[0.65rem] font-medium text-base-content/60 sm:inline-flex">
									<svg xmlns="http://www.w3.org/2000/svg" class="size-3 shrink-0 text-primary" viewBox="0 0 24 24"
									     fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
										<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/>
										<path d="m9 12 2 2 4-4"/>
									</svg>
									из коробки
								</span>
							</div>
							<p class="shrink-0 text-xs text-base-content/50">
								<span class="font-semibold text-base-content tabular-nums">1&nbsp;285</span>
								записей
							</p>
						</div>

						<!-- таблица -->
						<div class="px-2 pb-2">
							<div class="grid grid-cols-[1fr_auto_auto] items-center gap-x-3 px-3 pb-2 text-[0.7rem] font-medium uppercase tracking-wide text-base-content/40">
								<span>Клиент</span>
								<span class="text-right">Статус</span>
								<span class="text-right">Сумма</span>
							</div>

							<ul class="space-y-0.5 text-sm">
								<li class="zv-rise zv-d1 grid grid-cols-[1fr_auto_auto] items-center gap-x-3 rounded-lg px-3 py-2.5">
									<div class="flex min-w-0 items-center gap-2.5">
										<span class="grid size-7 shrink-0 place-items-center rounded-full bg-base-200 text-xs font-semibold text-base-content/70">АК</span>
										<span class="truncate text-base-content">Анна Ковалёва</span>
									</div>
									<span class="badge badge-sm badge-success badge-soft">Активен</span>
									<span class="text-right font-medium tabular-nums text-base-content">₽ 48 200</span>
								</li>

								<!-- строка «только что обновлена» -->
								<li class="zv-rise zv-d2 zv-row-live grid grid-cols-[1fr_auto_auto] items-center gap-x-3 rounded-lg px-3 py-2.5 ring-1 ring-inset ring-primary/20">
									<div class="flex min-w-0 items-center gap-2.5">
										<span class="grid size-7 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-semibold text-primary">ДМ</span>
										<span class="truncate font-medium text-base-content">Дмитрий Марков</span>
										<span class="zv-row-mark hidden items-center rounded-full bg-primary px-1.5 py-0.5 text-[0.6rem] font-semibold leading-none text-primary-content sm:inline-flex">
											обновлено
										</span>
									</div>
									<span class="badge badge-sm badge-warning badge-soft">В работе</span>
									<span class="text-right font-semibold tabular-nums text-base-content">₽ 12 900</span>
								</li>

								<li class="zv-rise zv-d3 grid grid-cols-[1fr_auto_auto] items-center gap-x-3 rounded-lg px-3 py-2.5">
									<div class="flex min-w-0 items-center gap-2.5">
										<span class="grid size-7 shrink-0 place-items-center rounded-full bg-base-200 text-xs font-semibold text-base-content/70">ЕС</span>
										<span class="truncate text-base-content">Елена Соколова</span>
									</div>
									<span class="badge badge-sm badge-ghost">Черновик</span>
									<span class="text-right font-medium tabular-nums text-base-content">₽ 5 400</span>
								</li>

								<li class="zv-rise zv-d4 grid grid-cols-[1fr_auto_auto] items-center gap-x-3 rounded-lg px-3 py-2.5">
									<div class="flex min-w-0 items-center gap-2.5">
										<span class="grid size-7 shrink-0 place-items-center rounded-full bg-base-200 text-xs font-semibold text-base-content/70">ПН</span>
										<span class="truncate text-base-content">Павел Новиков</span>
									</div>
									<span class="badge badge-sm badge-success badge-soft">Активен</span>
									<span class="text-right font-medium tabular-nums text-base-content">₽ 91 750</span>
								</li>
							</ul>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Преимущества шаблона. НЕ три одинаковые карточки-близнецы: крупное
		     утверждение слева + список выгод справа (ритм, не сетка). Каждая выгода —
		     на языке пользы, техника скрыта. -->
		<section class="border-t border-base-300 bg-base-200/40">
			<div class="mx-auto w-full max-w-6xl px-6 py-20 lg:px-8">
				<div class="grid gap-12 lg:grid-cols-[1fr_1.4fr]">
					<div>
						<h2 class="font-display text-3xl font-bold tracking-tight text-balance">
							Крепкий фундамент, а не пустой лист
						</h2>
						<p class="mt-4 text-base-content/70">
							Скучное и обязательное уже сделано и проверено. Остаётся приятная
							часть — собрать то, ради чего приложение и создаётся.
						</p>
					</div>
					<div class="grid gap-x-10 gap-y-8 sm:grid-cols-2">
						<div>
							<h3 class="font-semibold">Аккаунты из коробки</h3>
							<p class="mt-1.5 text-sm text-base-content/70">
								Вход, регистрация, восстановление пароля, подтверждение почты и
								роли — готово и безопасно, ничего не надо настраивать.
							</p>
						</div>
						<div>
							<h3 class="font-semibold">Панель управления данными</h3>
							<p class="mt-1.5 text-sm text-base-content/70">
								Удобный бэкофис для ваших записей — список, форма, поиск и
								фильтры появляются для каждого раздела автоматически.
							</p>
						</div>
						<div>
							<h3 class="font-semibold">Мгновенный отклик</h3>
							<p class="mt-1.5 text-sm text-base-content/70">
								Страницы не перезагружаются и не мигают — приложение
								откликается сразу, как хорошее настольное.
							</p>
						</div>
						<div>
							<h3 class="font-semibold">Оформление на любой вкус</h3>
							<p class="mt-1.5 text-sm text-base-content/70">
								Десятки готовых тем — светлые, тёмные, яркие. Переключаются в
								один клик, приложение подстраивается целиком.
							</p>
						</div>
						<div>
							<h3 class="font-semibold">Супербыстрое и выносливое</h3>
							<p class="mt-1.5 text-sm text-base-content/70">
								Работает быстро и держит нагрузку — от первых пользователей до
								наплыва, без переделок под ростом.
							</p>
						</div>
						<div>
							<h3 class="font-semibold">Растёт вместе с идеей</h3>
							<p class="mt-1.5 text-sm text-base-content/70">
								Добавить новый раздел — дело минут. Основа устроена так, чтобы
								приложение развивалось, не спотыкаясь.
							</p>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Финальный призыв: одно действие, крупно, осознанный центр. -->
		<section class="mx-auto w-full max-w-3xl px-6 py-20 text-center lg:px-8">
			<h2 class="font-display text-3xl font-bold tracking-tight text-balance sm:text-4xl">
				Начните с готового — доведите до своего
			</h2>
			<p class="mx-auto mt-4 max-w-md text-base-content/70">
				Создайте аккаунт за минуту и посмотрите, как это работает.
			</p>
			<div class="mt-8">
				{{if .User}}
				<a class="btn btn-primary btn-lg" href="/">Открыть приложение</a>
				{{else}}
				<a class="btn btn-primary btn-lg" href="/register">Создать аккаунт</a>
				{{end}}
			</div>
		</section>
	</main>

	{{template "site-footer" .}}
</div>
{{end}}
```

## `internal/transport/web/templates/layout.html`

```html
{{define "layout"}}<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<!-- Заголовок вкладки: «Страница · Приложение». AppName — имя приложения,
	     Title — конкретная страница. На лендинге Title пуст → только имя. -->
	<title>{{if .Title}}{{.Title}} · {{end}}{{if .AppName}}{{.AppName}}{{else}}Приложение{{end}}</title>

	<!-- Стили приложения — Tailwind CSS + DaisyUI, собраны в static/app.css
	     (см. assets/input.css, `make css`). Локальная статика, без CDN. Верстай
	     на компонентах DaisyUI (см. skill daisyui). -->
	<link rel="stylesheet" href="/static/app.css">

	<!-- htmx для интерактива (локально, без CDN — важно для работы в РФ). -->
	<script src="/static/htmx.min.js"></script>

	<!-- Тема оформления. daisyUI даёт множество встроенных тем (см. список в
	     assets/input.css). По умолчанию (пользователь ничего не выбирал) тему
	     задаёт система: светлая, а при тёмной системной настройке — тёмная
	     (--default / --prefersdark в input.css). Как только пользователь выбрал
	     тему переключателем в шапке — её имя сохраняется в localStorage и ставится
	     атрибутом data-theme на <html>, перекрывая системную. Скрипт стоит в
	     <head> и применяет сохранённую тему ДО отрисовки, чтобы не было вспышки
	     чужой темы. hx-boost меняет только <body>, поэтому <html> и этот скрипт
	     живут всю сессию; выбранный пункт в меню синхронизируется после swap. -->
	<script>
		(function () {
			// Прочитать сохранённое имя темы (или null, если выбора не было).
			function stored() {
				try { return localStorage.getItem("theme"); } catch (e) { return null; }
			}

			// Применить сохранённый выбор к <html>. Нет выбора — снимаем data-theme,
			// тогда работает системная prefers-color-scheme (light/dark по системе).
			function applyStored() {
				var t = stored();
				if (t) document.documentElement.setAttribute("data-theme", t);
				else document.documentElement.removeAttribute("data-theme");
			}
			applyStored();

			// Отметить в меню радио-пункт активной темы (после hx-boost и на старте).
			function syncRadios() {
				var t = stored();
				document.querySelectorAll("input.theme-controller").forEach(function (el) {
					el.checked = !!t && el.value === t;
				});
			}

			// Выбор темы в меню: сохранить имя и применить сразу.
			document.addEventListener("change", function (e) {
				var el = e.target;
				if (!el.classList || !el.classList.contains("theme-controller")) return;
				if (!el.value) return;
				try { localStorage.setItem("theme", el.value); } catch (err) {}
				document.documentElement.setAttribute("data-theme", el.value);
			});

			document.addEventListener("htmx:afterSwap", syncRadios);
			document.addEventListener("DOMContentLoaded", syncRadios);
		})();
	</script>
</head>
<!-- hx-boost: приложение работает как SPA. htmx перехватывает клики по обычным
     ссылкам <a> и сабмиты форм, грузит страницу в фоне и подменяет <body> БЕЗ
     полной перезагрузки — контент не пропадает на миллисекунду, нет «мигания».
     URL и история браузера обновляются (кнопка «назад» работает). Держи этот
     атрибут — это и есть плавная навигация без морганий. Тема DaisyUI (светлая
     по умолчанию, тёмная по системной настройке) задана в assets/input.css. -->
<body hx-boost="true" class="min-h-screen bg-base-100 text-base-content">
	{{if eq .Page "landing"}}
	<!-- Лендинг — полноэкранный, без стандартной шапки. -->
	{{template "landing" .}}
	{{else}}
	<!-- Внутренние страницы и страницы входа: та же шапка сайта (логотип-ссылка на
	     главную → возврат откуда угодно). Футер прижат книзу через flex-колонку. -->
	<div class="flex min-h-screen flex-col">
		{{template "site-header" .}}
		<main class="mx-auto w-full max-w-3xl flex-1 px-4 py-10 md:px-8">
			{{if eq .Page "login"}}{{template "login" .}}
			{{else if eq .Page "register"}}{{template "register" .}}
			{{else if eq .Page "forgot"}}{{template "forgot" .}}
			{{else if eq .Page "reset"}}{{template "reset" .}}
			{{else if eq .Page "verify"}}{{template "verify" .}}
			{{else if eq .Page "settings"}}{{template "settings" .}}
			{{else}}{{template "content" .}}{{end}}
		</main>
		{{template "site-footer" .}}
	</div>
	{{end}}
</body>
</html>{{end}}

<!-- Переключатель темы: выпадающее меню (dropdown) со списком всех встроенных
     тем daisyUI. Каждый пункт — radio с классом theme-controller (компонент
     daisyUI): выбор радио мгновенно применяет тему к странице; рядом мини-превью
     из 4 цветов темы (обёртка data-theme="имя" + bg-* берут цвета этой темы, без
     хардкода). Открытие/закрытие меню — на <details>/<summary>, без JS. Сохранение
     выбора и синхронизация активного пункта — скриптом в <head> layout. -->
{{define "theme-toggle"}}
<details class="dropdown dropdown-end">
	<summary class="btn btn-ghost btn-circle btn-sm" aria-label="Выбрать тему">
		<svg class="h-5 w-5 fill-current" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 3a9 9 0 0 0 0 18 2.5 2.5 0 0 0 2.5-2.5c0-.6-.2-1.1-.6-1.6-.3-.4-.5-.8-.5-1.3a1.5 1.5 0 0 1 1.5-1.5H16a5 5 0 0 0 5-5c0-4.4-4-8-9-8Zm-5.5 9a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm3-4a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Z"/></svg>
	</summary>
	<ul class="dropdown-content menu z-10 mt-2 max-h-96 w-56 flex-nowrap overflow-y-auto rounded-box border border-base-300 bg-base-100 p-2 shadow-lg">
		<li class="menu-title">Тема оформления</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="light" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="light">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">light</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="dark" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="dark">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">dark</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="cupcake" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="cupcake">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">cupcake</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="bumblebee" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="bumblebee">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">bumblebee</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="emerald" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="emerald">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">emerald</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="corporate" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="corporate">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">corporate</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="synthwave" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="synthwave">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">synthwave</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="retro" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="retro">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">retro</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="cyberpunk" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="cyberpunk">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">cyberpunk</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="valentine" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="valentine">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">valentine</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="halloween" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="halloween">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">halloween</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="garden" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="garden">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">garden</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="forest" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="forest">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">forest</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="aqua" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="aqua">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">aqua</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="lofi" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="lofi">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">lofi</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="pastel" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="pastel">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">pastel</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="fantasy" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="fantasy">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">fantasy</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="wireframe" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="wireframe">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">wireframe</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="black" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="black">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">black</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="luxury" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="luxury">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">luxury</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="dracula" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="dracula">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">dracula</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="cmyk" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="cmyk">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">cmyk</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="autumn" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="autumn">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">autumn</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="business" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="business">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">business</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="acid" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="acid">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">acid</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="lemonade" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="lemonade">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">lemonade</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="night" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="night">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">night</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="coffee" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="coffee">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">coffee</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="winter" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="winter">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">winter</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="dim" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="dim">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">dim</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="nord" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="nord">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">nord</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="sunset" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="sunset">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">sunset</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="caramellatte" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="caramellatte">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">caramellatte</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="abyss" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="abyss">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">abyss</span>
				</label>
			</li>
			<li>
				<label class="flex items-center gap-3 px-2">
					<input type="radio" name="theme-dropdown" value="silk" class="theme-controller sr-only" />
					<span class="grid shrink-0 grid-cols-2 gap-0.5 rounded-md p-1 shadow-sm" data-theme="silk">
						<span class="size-1.5 rounded-full bg-base-content"></span>
						<span class="size-1.5 rounded-full bg-primary"></span>
						<span class="size-1.5 rounded-full bg-secondary"></span>
						<span class="size-1.5 rounded-full bg-accent"></span>
					</span>
					<span class="w-full truncate text-sm capitalize">silk</span>
				</label>
			</li>
	</ul>
</details>
{{end}}

<!-- Шапка сайта: логотип-ссылка на главную (возврат с любой страницы), навигация
     и вход/регистрация для гостя (или профиль для вошедшего) + переключатель тем.
     Общая для лендинга и страниц входа — поэтому с login/forgot/reset всегда можно
     вернуться на главную. Логотип = название приложения ({{.Title}}), агент задаёт
     его под конкретный продукт. Без упоминания технологий/стека. -->
{{define "site-header"}}
<header class="sticky top-0 z-30 border-b border-base-300 bg-base-100/80 backdrop-blur">
	<div class="mx-auto flex w-full max-w-6xl items-center gap-4 px-6 py-3 lg:px-8">
		<!-- Логотип: значок в плашке акцентного цвета + название приложения.
		     Иконка — inline SVG (локально, без CDN). Агент может заменить её на
		     реальный логотип продукта. Название — из настройки app_name (.AppName). -->
		<a href="/" class="flex items-center gap-2.5 text-lg font-bold tracking-tight">
			<span class="grid h-8 w-8 place-items-center rounded-lg bg-primary text-primary-content">
<svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor" stroke="none" aria-hidden="true">
					<!-- Значок электричества (молния, Lucide zap): цельная фигура,
					     чисто читается даже на мелком размере логотипа. -->
					<path d="M13 2 3 14h7l-1 8 10-12h-7z" />
				</svg>
			</span>
			<span>{{if .AppName}}{{.AppName}}{{else}}Zerovibe{{end}}</span>
		</a>
		<div class="flex-1"></div>
		<nav class="flex items-center gap-2">
			{{if .User}}
			<span class="hidden text-sm text-base-content/60 sm:inline">{{.User.Email}}</span>
			{{if eq .User.Role "admin"}}<a class="btn btn-ghost btn-sm" href="/admin">Админка</a>{{end}}
			<form method="post" action="/logout"><button class="btn btn-ghost btn-sm" type="submit">Выйти</button></form>
			{{else}}
			<a class="btn btn-ghost btn-sm" href="/login">Войти</a>
			<a class="btn btn-primary btn-sm" href="/register">Регистрация</a>
			{{end}}
			{{template "theme-toggle" .}}
		</nav>
	</div>
</header>
{{end}}

<!-- Подвал сайта: логотип, минимум ссылок. Без технической информации. Ссылки —
     плейсхолдеры под продукт. -->
{{define "site-footer"}}
<footer class="border-t border-base-300 bg-base-200/40">
	<div class="mx-auto flex w-full max-w-6xl flex-col items-center gap-3 px-6 py-8 text-sm text-base-content/60 sm:flex-row sm:justify-between lg:px-8">
		<span class="font-semibold text-base-content">{{if .AppName}}{{.AppName}}{{else}}Приложение{{end}}</span>
		<nav class="flex items-center gap-5">
			<a href="/login" class="link link-hover">Войти</a>
			<a href="/register" class="link link-hover">Регистрация</a>
		</nav>
	</div>
</footer>
{{end}}
```

## `internal/transport/web/templates/login.html`

```html
{{define "login"}}
<div class="mx-auto mt-6 w-full max-w-sm">
	<div class="card border border-base-300 bg-base-100">
		<div class="card-body gap-4">
			{{if .Flash}}<div class="alert alert-info text-sm">{{.Flash}}</div>{{end}}
			{{if .Err}}<div class="alert alert-error text-sm">{{.Err}}</div>{{end}}
			<form method="post" action="/login">
				<fieldset class="fieldset">
					<label class="label" for="email">Email</label>
					<input id="email" class="input w-full" type="email" name="email" autocomplete="username" required autofocus>
					<label class="label" for="password">Пароль</label>
					<input id="password" class="input w-full" type="password" name="password" autocomplete="current-password" required>
					<button class="btn btn-primary btn-block mt-2" type="submit">Войти</button>
				</fieldset>
			</form>
			<div class="flex justify-between text-sm">
				<a class="link link-hover" href="/forgot">Забыли пароль?</a>
				{{if .AllowSignup}}<a class="link link-hover" href="/register">Регистрация</a>{{end}}
			</div>
		</div>
	</div>
</div>
{{end}}
```

## `internal/transport/web/templates/note.html`

```html
{{define "note"}}<div class="card border border-base-300 bg-base-100" id="note-{{.ID}}">
	<div class="card-body gap-2">
		<div class="flex items-start justify-between gap-3">
			<h3 class="card-title text-base">{{.Title}}</h3>
			<!-- HTMX: DELETE возвращает пустоту, htmx убирает сам элемент (outerHTML). -->
			<button class="btn btn-ghost btn-xs text-error"
				hx-delete="/notes/{{.ID}}"
				hx-target="#note-{{.ID}}"
				hx-swap="outerHTML"
				hx-confirm="Удалить заметку?">Удалить</button>
		</div>
		{{if .Body}}<p class="whitespace-pre-wrap text-sm text-base-content/80">{{.Body}}</p>{{end}}
		{{if not .CreatedAt.IsZero}}<time class="text-xs text-base-content/50">{{.CreatedAt.Format "02.01.2006 15:04"}}</time>{{end}}
	</div>
</div>{{end}}
```

## `internal/transport/web/templates/notes.html`

```html
{{define "content"}}
	<!-- Форма создания. HTMX: POST возвращает фрагмент "note", вставляем в начало
	     #notes (afterbegin), форму сбрасываем после ответа. -->
	<div class="card mb-4 border border-base-300 bg-base-100">
		<div class="card-body">
			<form class="flex flex-col gap-3" hx-post="/notes" hx-target="#notes" hx-swap="afterbegin" hx-on::after-request="this.reset()">
				<input class="input w-full" name="title" placeholder="Заголовок" maxlength="200" required autofocus>
				<textarea class="textarea w-full" name="body" placeholder="Текст заметки (необязательно)" maxlength="10000" rows="3"></textarea>
				<button class="btn btn-primary self-start" type="submit">Добавить</button>
			</form>
		</div>
	</div>

	<div id="notes" class="flex flex-col gap-3">
		{{range .Notes}}{{template "note" .}}{{else}}<p class="py-8 text-center text-base-content/50" id="empty">Пока нет заметок. Добавьте первую.</p>{{end}}
	</div>
{{end}}
```

## `internal/transport/web/templates/register.html`

```html
{{define "register"}}
<div class="mx-auto mt-6 w-full max-w-sm">
	<div class="card border border-base-300 bg-base-100">
		<div class="card-body gap-4">
			{{if .AllowSignup}}
			{{if .Err}}<div class="alert alert-error text-sm">{{.Err}}</div>{{end}}
			<form method="post" action="/register">
				<fieldset class="fieldset">
					<label class="label" for="email">Email</label>
					<input id="email" class="input w-full" type="email" name="email" autocomplete="username" required autofocus>
					<label class="label" for="password">Пароль (минимум 8 символов)</label>
					<input id="password" class="input w-full" type="password" name="password" autocomplete="new-password" minlength="8" required>
					<button class="btn btn-primary btn-block mt-2" type="submit">Зарегистрироваться</button>
				</fieldset>
			</form>
			<div class="text-sm">
				<a class="link link-hover" href="/login">Уже есть аккаунт? Войти</a>
			</div>
			{{else}}
			<p class="text-sm text-base-content/70">Регистрация сейчас закрыта. Обратитесь к администратору приложения.</p>
			<div class="text-sm">
				<a class="link link-hover" href="/login">Войти</a>
			</div>
			{{end}}
		</div>
	</div>
</div>
{{end}}
```

## `internal/transport/web/templates/reset.html`

```html
{{define "reset"}}
<div class="mx-auto mt-6 w-full max-w-sm">
	<div class="card border border-base-300 bg-base-100">
		<div class="card-body gap-4">
			{{if .Token}}
			{{if .Err}}<div class="alert alert-error text-sm">{{.Err}}</div>{{end}}
			<p class="text-sm text-base-content/70">Задайте новый пароль.</p>
			<form method="post" action="/reset">
				<input type="hidden" name="token" value="{{.Token}}">
				<fieldset class="fieldset">
					<label class="label" for="password">Новый пароль (минимум 8 символов)</label>
					<input id="password" class="input w-full" type="password" name="password" autocomplete="new-password" minlength="8" required autofocus>
					<button class="btn btn-primary btn-block mt-2" type="submit">Сохранить пароль</button>
				</fieldset>
			</form>
			{{else}}
			<div class="alert alert-error text-sm">Ссылка недействительна. Запросите сброс заново.</div>
			<div class="text-sm"><a class="link link-hover" href="/forgot">Восстановить пароль</a></div>
			{{end}}
		</div>
	</div>
</div>
{{end}}
```

## `internal/transport/web/templates/settings.html`

```html
{{define "settings"}}
	<div class="card border border-base-300 bg-base-100">
		<div class="card-body gap-4">
			<p class="text-sm text-base-content/70">Настройки приложения. Секреты не отображаются — только признак «задано».</p>
			<div class="overflow-x-auto">
				<table class="table">
					<thead><tr><th>Настройка</th><th>Значение</th></tr></thead>
					<tbody>
					{{range .Settings}}
						<tr>
							<td>{{.Title}} <span class="text-base-content/50">({{.Key}})</span></td>
							<td>
								{{if eq .Kind "secret"}}
									{{if .Set}}<span class="badge badge-ghost">задано</span>{{else}}<span class="text-base-content/40">—</span>{{end}}
								{{else}}
									{{if .Value}}{{.Value}}{{else}}<span class="text-base-content/40">—</span>{{end}}
								{{end}}
							</td>
						</tr>
					{{end}}
					</tbody>
				</table>
			</div>
		</div>
	</div>
{{end}}
```

## `internal/transport/web/templates/verify.html`

```html
{{define "verify"}}
<div class="mx-auto mt-6 w-full max-w-sm">
	<div class="card border border-base-300 bg-base-100">
		<div class="card-body gap-4">
			{{if .Flash}}<div class="alert alert-info text-sm">{{.Flash}}</div>{{end}}
			{{if .Err}}<div class="alert alert-error text-sm">{{.Err}}</div>{{end}}
			<p class="text-sm text-base-content/70">Не пришло письмо? Отправим ещё раз.</p>
			<form method="post" action="/resend-verification">
				<fieldset class="fieldset">
					<label class="label" for="email">Email</label>
					<input id="email" class="input w-full" type="email" name="email" value="{{.Email}}" autocomplete="username" required>
					<button class="btn btn-primary btn-block mt-2" type="submit">Отправить письмо повторно</button>
				</fieldset>
			</form>
			<div class="text-sm"><a class="link link-hover" href="/login">Вернуться ко входу</a></div>
		</div>
	</div>
</div>
{{end}}
```

## `internal/transport/web/web.go`

```go
// Package web — HTTP/HTML транспорт на HTMX. Зависит от usecase и domain.
//
// КЛЮЧЕВОЙ HTMX-ПАТТЕРН:
//   - GET страницы — отдаёт ПОЛНУЮ страницу (layout + нужный content по .Page).
//   - Мутации (POST/PUT/DELETE) возвращают РОВНО изменившийся фрагмент, либо при
//     навигации просят клиента сделать редирект заголовком HX-Redirect.
//
// АУТЕНТИФИКАЦИЯ:
//   - loadUser — мягкое middleware: кладёт текущего пользователя в контекст (гость
//     проходит дальше без пользователя).
//   - requireAuth/requireRole — защита маршрутов (гость → на /login; не та роль → 403).
package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/chudno/zerovibe/internal/admin"
	"github.com/chudno/zerovibe/internal/domain"
	"github.com/chudno/zerovibe/internal/usecase"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// devTemplatesGlob / devStaticDir — пути шаблонов и статики на диске (относительно
// корня проекта) для dev-режима. В проде не используются — там всё из embed.
const (
	devTemplatesGlob = "internal/transport/web/templates/*.html"
	devStaticDir     = "internal/transport/web/static"
)

// Config — транспортный конфиг (поведение, не бизнес-правила).
type Config struct {
	SecureCookie bool   // ставить ли cookie с флагом Secure (true за TLS-edge; локально false)
	CookieName   string // имя cookie сессии (напр. "zv_session")
	// PreviewMode — приложение открыто в live-превью платформы (внутри cross-site
	// iframe по https). Тогда сессионная cookie ставится SameSite=None; Secure,
	// иначе браузер не сохранит её во фрейме и вход не удержится. В обычном
	// (задеплоенном) режиме — false, cookie остаётся SameSite=Lax (CSRF-защита).
	PreviewMode bool
}

// appName — название приложения для логотипа шапки/подвала и заголовка вкладки.
// Плейсхолдер эталона: агент задаёт его под конкретный продукт вайбкодера (замени
// строку здесь). Отдельно от pageData.Title (тот — заголовок КОНКРЕТНОЙ страницы,
// напр. «Вход»), чтобы в логотипе всегда было имя приложения, а не «Вход».
const appName = "Zerovibe"

// pageData — данные для рендера страниц. Page выбирает, какой content показать.
type pageData struct {
	AppName     string // имя приложения (логотип/подвал); проставляется в renderPage
	Title       string
	Page        string // "landing" | "notes" | "login" | "register" | "forgot" | "reset" | "settings"
	User        *domain.User
	Notes       []domain.Note
	Settings    []usecase.SettingView
	Flash       string // нейтральное сообщение (forgot/reset/verify)
	Err         string // текст ошибки формы
	Token       string // для формы reset
	Email       string // для формы повторной отправки подтверждения
	AllowSignup bool
}

// Server держит зависимости транспорта.
type Server struct {
	tmpl     *template.Template
	notes    *usecase.NoteService
	auth     *usecase.AuthService
	settings *usecase.SettingsService
	admin    *admin.Server // встроенная админка (nil → не смонтирована)
	cfg      Config
	dev      bool // dev-режим: шаблоны и статика читаются с диска (см. templates()/статику)
}

// templates возвращает набор шаблонов для рендера. В ПРОДЕ — вшитые в бинарь через
// embed (распарсены один раз в NewServer). В DEV-режиме (ZV_DEV=1) — перечитывает
// html-шаблоны с диска на КАЖДЫЙ рендер, чтобы правки вёрстки были видны сразу, без
// пересборки бинаря (embed фиксирует содержимое на этапе компиляции). Если чтение с
// диска не удалось (запуск из другого каталога) — молча откатываемся на вшитые.
func (s *Server) templates() *template.Template {
	if !s.dev {
		return s.tmpl
	}
	t, err := template.ParseGlob(devTemplatesGlob)
	if err != nil {
		return s.tmpl
	}
	return t
}

// SetAdmin подключает встроенную админку. Её маршруты монтируются в Routes() под
// защитой роли администратора. nil/пустой реестр → админка не появляется.
func (s *Server) SetAdmin(a *admin.Server) { s.admin = a }

// NewServer парсит шаблоны и собирает сервер.
func NewServer(notes *usecase.NoteService, auth *usecase.AuthService, settings *usecase.SettingsService, cfg Config) (*Server, error) {
	if cfg.CookieName == "" {
		cfg.CookieName = "zv_session"
	}
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	// ZV_DEV=1 включает live-reload вёрстки (шаблоны/статика с диска). В проде НЕ
	// ставится — приложение работает целиком из embed. См. templates() и Routes().
	dev := os.Getenv("ZV_DEV") == "1"
	return &Server{tmpl: tmpl, notes: notes, auth: auth, settings: settings, cfg: cfg, dev: dev}, nil
}

// Routes возвращает http.Handler со всеми маршрутами, обёрнутыми в loadUser.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Публичные (аутентификация).
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.HandleFunc("POST /logout", s.handleLogout)
	mux.HandleFunc("GET /register", s.handleRegisterPage)
	mux.HandleFunc("POST /register", s.handleRegister)
	mux.HandleFunc("GET /forgot", s.handleForgotPage)
	mux.HandleFunc("POST /forgot", s.handleForgot)
	mux.HandleFunc("GET /reset", s.handleResetPage)
	mux.HandleFunc("POST /reset", s.handleReset)
	mux.HandleFunc("GET /verify-email", s.handleVerifyEmail)
	mux.HandleFunc("POST /resend-verification", s.handleResendVerification)

	// Первичная настройка: создать первого администратора по одноразовому коду.
	// Доступна только пока админов нет; код печатается в лог при первом старте.
	mux.HandleFunc("POST /setup", s.handleSetup)

	// Служебное и статика.
	mux.HandleFunc("GET /healthz", s.handleHealth)
	// В проде статика (в т.ч. собранный app.css) — из embed. В dev-режиме отдаём с
	// диска, чтобы пересобранный `make css` был виден без пересборки бинаря.
	if s.dev {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(devStaticDir))))
	} else {
		mux.Handle("GET /static/", http.FileServerFS(staticFS))
	}

	// Публичная главная — лендинг «эталонный шаблон» (без авторизации).
	mux.HandleFunc("GET /", s.handleLanding)

	// Защищённые (демо-раздел заметок — личные, эталон сущности для новых фич).
	mux.HandleFunc("GET /notes", s.requireAuth(s.handleIndex))
	mux.HandleFunc("POST /notes", s.requireAuth(s.handleCreate))
	mux.HandleFunc("DELETE /notes/{id}", s.requireAuth(s.handleDelete))

	// Админ (настройки приложения).
	mux.HandleFunc("GET /admin/settings", s.requireRole(domain.RoleAdmin, s.handleSettingsPage))
	mux.HandleFunc("PUT /admin/settings", s.requireRole(domain.RoleAdmin, s.handleSetSetting))

	// Встроенная админка (CRUD над сущностями). У неё ОТДЕЛЬНЫЙ вход /admin/login со
	// своим дизайном (те же учётки, роль admin). Гость/не-админ на /admin/* → редирект
	// на /admin/login (а не на общий /login приложения). Все CRUD-маршруты — под guard.
	if s.admin != nil && s.admin.HasResources() {
		mux.HandleFunc("GET /admin/login", s.handleAdminLoginPage)
		mux.HandleFunc("POST /admin/login", s.handleAdminLogin)
		mux.HandleFunc("POST /admin/logout", s.handleAdminLogout)
		s.admin.Mount(mux, s.requireAdmin)
	}

	return s.loadUser(mux)
}

// --- контекст текущего пользователя ---

type ctxKey int

const userKey ctxKey = 0

// currentUser достаёт пользователя из контекста (nil если гость).
func currentUser(r *http.Request) *domain.User {
	u, _ := r.Context().Value(userKey).(*domain.User)
	return u
}

// --- middleware ---

// loadUser читает cookie сессии и кладёт пользователя в контекст. Гость проходит
// дальше без пользователя — это «мягкая» аутентификация для всех маршрутов.
func (s *Server) loadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(s.cfg.CookieName)
		if err == nil && c.Value != "" {
			if u, err := s.auth.Authenticate(r.Context(), c.Value); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), userKey, &u))
			} else if errors.Is(err, domain.ErrUnauthenticated) {
				// сессия истекла/недействительна — чистим cookie
				s.clearSessionCookie(w)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// requireAuth пропускает только аутентифицированных; гостя отправляет на вход.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r) == nil {
			s.fail(w, r, domain.ErrUnauthenticated)
			return
		}
		next(w, r)
	}
}

// requireRole пропускает только пользователей с нужной ролью.
func (s *Server) requireRole(role domain.Role, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := currentUser(r)
		if u == nil {
			s.fail(w, r, domain.ErrUnauthenticated)
			return
		}
		if u.Role != role {
			s.fail(w, r, domain.ErrForbidden)
			return
		}
		next(w, r)
	}
}

// requireAdmin — guard для встроенной админки: пускает только админов, а гостя/не-админа
// отправляет на ОТДЕЛЬНЫЙ вход /admin/login (не на общий /login приложения). Для htmx
// делает это через HX-Redirect, чтобы переход случился без «застрявшего» фрагмента.
func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u := currentUser(r); u != nil && u.Role == domain.RoleAdmin {
			next(w, r)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/admin/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	}
}

// --- хендлеры: страницы ---

// handleIndex — полная страница со списком личных заметок.
// handleLanding — публичная главная (лендинг «эталонный шаблон»). Точный роут
// "GET /"; ServeMux уводит неизвестные пути в NotFound автоматически, поэтому
// ручная проверка пути тут не нужна. Залогиненного пользователя мягко ведём в
// его раздел (эталон заметок), гостю показываем лендинг.
func (s *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.renderPage(w, r, pageData{Page: "landing", User: currentUser(r)})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	notes, err := s.notes.List(r.Context(), u.ID)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	s.renderPage(w, r, pageData{Title: "Заметки", Page: "notes", User: u, Notes: notes})
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if u := currentUser(r); u != nil {
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
		return
	}
	allow, _ := s.settings.Bool(r.Context(), "allow_signup")
	s.renderPage(w, r, pageData{Title: "Вход", Page: "login", AllowSignup: allow})
}

func (s *Server) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	if u := currentUser(r); u != nil {
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
		return
	}
	allow, _ := s.settings.Bool(r.Context(), "allow_signup")
	s.renderPage(w, r, pageData{Title: "Регистрация", Page: "register", AllowSignup: allow})
}

func (s *Server) handleForgotPage(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, r, pageData{Title: "Восстановление пароля", Page: "forgot"})
}

func (s *Server) handleResetPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	s.renderPage(w, r, pageData{Title: "Новый пароль", Page: "reset", Token: token})
}

func (s *Server) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	views, err := s.settings.All(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	s.renderPage(w, r, pageData{Title: "Настройки", Page: "settings", User: currentUser(r), Settings: views})
}

// --- хендлеры: мутации ---

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	rateKey := domain.NormalizeEmail(email) + "|" + clientIP(r)

	sess, err := s.auth.Login(r.Context(), email, password, rateKey)
	if err != nil {
		// Почта не подтверждена — показываем страницу подтверждения с кнопкой повтора.
		if errors.Is(err, domain.ErrEmailNotVerified) {
			s.renderPage(w, r, pageData{
				Title: "Подтвердите почту", Page: "verify",
				Email: domain.NormalizeEmail(email),
				Flash: "Аккаунт создан, но почта ещё не подтверждена. Перейдите по ссылке из письма.",
			})
			return
		}
		s.failForm(w, r, "login", err)
		return
	}
	s.setSessionCookie(w, sess)
	s.redirect(w, r, "/notes")
}

// handleAdminLoginPage показывает отдельную форму входа в админку (свой дизайн).
// Если уже вошёл админом — сразу в /admin.
func (s *Server) handleAdminLoginPage(w http.ResponseWriter, r *http.Request) {
	if u := currentUser(r); u != nil && u.Role == domain.RoleAdmin {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	s.admin.RenderLogin(w, "", "")
}

// handleAdminLogin — вход в админку: те же учётки, что в приложении, но пускаем ТОЛЬКО
// роль admin. Успех → сессия + переход в /admin (через HX-Redirect, форма на htmx).
// Не админ или неверные креды → форма входа админки с ошибкой (без перезагрузки).
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	rateKey := domain.NormalizeEmail(email) + "|" + clientIP(r)

	sess, err := s.auth.Login(r.Context(), email, password, rateKey)
	if err != nil {
		s.admin.RenderLogin(w, domain.NormalizeEmail(email), "Неверный email или пароль.")
		return
	}
	// Сессия создана — проверяем роль. Не админ: не пускаем в админку (но в приложении
	// его сессия валидна — кладём cookie и отправляем в приложение, без админ-доступа).
	u, uerr := s.auth.Authenticate(r.Context(), sess.Token)
	if uerr != nil || u.Role != domain.RoleAdmin {
		s.admin.RenderLogin(w, domain.NormalizeEmail(email), "У этой учётной записи нет доступа к админке.")
		return
	}
	s.setSessionCookie(w, sess)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// handleAdminLogout выходит из админки и возвращает на её вход.
func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.cfg.CookieName); err == nil {
		_ = s.auth.Logout(r.Context(), c.Value)
	}
	s.clearSessionCookie(w)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/login")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// handleVerifyEmail подтверждает почту по токену из ссылки в письме.
func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if err := s.auth.ConfirmEmailVerification(r.Context(), token); err != nil {
		s.renderPage(w, r, pageData{Title: "Подтверждение почты", Page: "verify",
			Err: "Ссылка недействительна или устарела. Запросите письмо повторно."})
		return
	}
	s.renderPage(w, r, pageData{Title: "Вход", Page: "login",
		Flash: "Почта подтверждена. Теперь можно войти."})
}

// handleResendVerification повторно отправляет письмо подтверждения (рейт-лимит).
func (s *Server) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	rateKey := domain.NormalizeEmail(email) + "|" + clientIP(r)
	if err := s.auth.ResendVerification(r.Context(), email, rateKey); err != nil {
		// Рейт-лимит — единственная ошибка наружу (анти-enumeration в сервисе).
		var limited domain.ErrRateLimited
		if errors.As(err, &limited) {
			w.Header().Set("Retry-After", strconv.Itoa(int(limited.RetryAfter.Seconds())))
		}
		s.renderPage(w, r, pageData{Title: "Подтвердите почту", Page: "verify",
			Email: domain.NormalizeEmail(email), Err: errText(err)})
		return
	}
	s.renderPage(w, r, pageData{Title: "Подтвердите почту", Page: "verify",
		Email: domain.NormalizeEmail(email),
		Flash: "Если адрес ещё не подтверждён, мы отправили новое письмо. Проверьте почту."})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.cfg.CookieName); err == nil && c.Value != "" {
		_ = s.auth.Logout(r.Context(), c.Value)
	}
	s.clearSessionCookie(w)
	s.redirect(w, r, "/login")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if _, err := s.auth.Register(r.Context(), email, password); err != nil {
		s.failForm(w, r, "register", err)
		return
	}
	// Автологин после регистрации.
	rateKey := domain.NormalizeEmail(email) + "|" + clientIP(r)
	sess, err := s.auth.Login(r.Context(), email, password, rateKey)
	if err != nil {
		// Регистрация прошла, но автологин не удался (напр. включено подтверждение
		// почты → ErrEmailNotVerified) — отправляем на вход.
		s.redirect(w, r, "/login")
		return
	}
	s.setSessionCookie(w, sess)
	s.redirect(w, r, "/notes")
}

// handleSetup — первичная настройка: создаёт ПЕРВОГО администратора по одноразовому
// коду (печатается в лог при первом старте). Это служебный API-эндпоинт, который
// дёргает агент после деплоя, поэтому отвечает простым текстом + кодом, а не страницей.
// Принимает поля email/password/token (form или JSON-тело).
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	email, password, token := r.FormValue("email"), r.FormValue("password"), r.FormValue("token")
	// Допускаем и JSON-тело (агенту удобнее).
	if email == "" && password == "" && token == "" {
		var body struct{ Email, Password, Token string }
		if json.NewDecoder(r.Body).Decode(&body) == nil {
			email, password, token = body.Email, body.Password, body.Token
		}
	}

	if err := s.auth.Setup(r.Context(), email, password, token); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("администратор создан, можно входить\n"))
}

func (s *Server) handleForgot(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	rateKey := domain.NormalizeEmail(email) + "|" + clientIP(r)

	if err := s.auth.RequestReset(r.Context(), email, rateKey); err != nil {
		s.failForm(w, r, "forgot", err)
		return
	}
	// Анти-enumeration: всегда нейтральный ответ.
	s.renderPage(w, r, pageData{
		Title: "Восстановление пароля", Page: "forgot",
		Flash: "Если такой email зарегистрирован, мы отправили на него ссылку для сброса пароля.",
	})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	password := r.FormValue("password")

	if err := s.auth.ConfirmReset(r.Context(), token, password); err != nil {
		s.renderPage(w, r, pageData{Title: "Новый пароль", Page: "reset", Token: token, Err: errText(err)})
		return
	}
	s.renderPage(w, r, pageData{
		Title: "Вход", Page: "login",
		Flash: "Пароль изменён. Войдите с новым паролем.",
	})
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	n, err := s.notes.Create(r.Context(), u.ID, r.FormValue("title"), r.FormValue("body"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates().ExecuteTemplate(w, "note", n); err != nil {
		s.fail(w, r, err)
	}
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "некорректный id", http.StatusBadRequest)
		return
	}
	u := currentUser(r)
	if err := s.notes.Delete(r.Context(), id, u.ID); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSetSetting(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")
	if err := s.settings.Set(r.Context(), key, value); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// --- рендер и ошибки ---

func (s *Server) renderPage(w http.ResponseWriter, r *http.Request, data pageData) {
	if data.User == nil {
		data.User = currentUser(r)
	}
	// Название приложения — из настройки app_name (её вайбкодер меняет в админке
	// без правки кода). Пусто/ошибка → константа appName как фолбэк.
	data.AppName = appName
	if s.settings != nil {
		if name, err := s.settings.String(r.Context(), "app_name"); err == nil && name != "" {
			data.AppName = name
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates().ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
	}
}

// failForm перерисовывает страницу-форму с текстом ошибки (для login/register/forgot).
func (s *Server) failForm(w http.ResponseWriter, r *http.Request, page string, err error) {
	// Рейт-лимит — отдельный код и заголовок, даже для форм.
	var limited domain.ErrRateLimited
	if errors.As(err, &limited) {
		w.Header().Set("Retry-After", strconv.Itoa(int(limited.RetryAfter.Seconds())))
	}
	allow, _ := s.settings.Bool(r.Context(), "allow_signup")
	title := map[string]string{"login": "Вход", "register": "Регистрация", "forgot": "Восстановление пароля"}[page]
	s.renderPage(w, r, pageData{Title: title, Page: page, Err: errText(err), AllowSignup: allow})
}

// fail мапит доменные ошибки в HTTP-коды. Единая точка обработки ошибок транспорта.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	var notFound domain.ErrNotFound
	var validation domain.ErrValidation
	var limited domain.ErrRateLimited
	switch {
	case errors.As(err, &notFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.As(err, &validation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrSignupClosed):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrEmailTaken):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrInvalidToken):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrEmailNotVerified):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrSetupClosed):
		http.Error(w, err.Error(), http.StatusGone)
	case errors.Is(err, domain.ErrSetupToken):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrUnauthenticated):
		if isHTMX(r) {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	case errors.As(err, &limited):
		w.Header().Set("Retry-After", strconv.Itoa(int(limited.RetryAfter.Seconds())))
		http.Error(w, err.Error(), http.StatusTooManyRequests)
	default:
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
	}
}

// errText возвращает безопасный для показа пользователю текст ошибки.
func errText(err error) string {
	var validation domain.ErrValidation
	if errors.As(err, &validation) {
		return validation.Error()
	}
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrSignupClosed),
		errors.Is(err, domain.ErrEmailTaken),
		errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, domain.ErrEmailNotVerified):
		return err.Error()
	}
	var limited domain.ErrRateLimited
	if errors.As(err, &limited) {
		return limited.Error()
	}
	return "что-то пошло не так, попробуйте ещё раз"
}

// --- cookie и вспомогательные ---

// Защита от CSRF: cookie сессии помечена SameSite=Lax — браузер не отправляет её при
// межсайтовых POST-запросах, что закрывает классический CSRF на мутации. Отдельных
// CSRF-токенов нет: для приложения такого класса (формы того же origin, Lax-cookie)
// это осознанное упрощение. Если понадобится строже — добавить токен в формы.
//
// Исключение — live-превью платформы (PreviewMode): приложение открыто в cross-site
// iframe по https, куда Lax-cookie браузер не пускает. Тогда ставим SameSite=None;
// Secure — единственный способ удержать сессию во фрейме (после регистрации/входа).
// Ослабление CSRF здесь приемлемо: превью — эфемерная песочница, доступная только
// владельцу через подписанный доступ платформы, не публичный прод.
func (s *Server) setSessionCookie(w http.ResponseWriter, sess domain.Session) {
	c := &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    sess.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  sess.ExpiresAt,
	}
	s.applyCookieSiteMode(c)
	http.SetCookie(w, c)
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	s.applyCookieSiteMode(c)
	http.SetCookie(w, c)
}

// applyCookieSiteMode в режиме превью переключает cookie на SameSite=None; Secure
// (нужно для cross-site iframe). SameSite=None без Secure браузер отвергает, поэтому
// Secure выставляется принудительно — превью всегда за https-edge платформы.
func (s *Server) applyCookieSiteMode(c *http.Cookie) {
	if s.cfg.PreviewMode {
		c.SameSite = http.SameSiteNoneMode
		c.Secure = true
	}
}

// redirect делает навигацию: для htmx — заголовком HX-Redirect, иначе 303.
func (s *Server) redirect(w http.ResponseWriter, r *http.Request, to string) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", to)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
}

// isHTMX сообщает, пришёл ли запрос от htmx.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// clientIP определяет IP клиента для ключа рейт-лимита.
//
// БЕЗОПАСНОСТЬ: НЕ доверяем X-Forwarded-For — этот заголовок полностью подделывается
// клиентом, и доверие к нему позволяет обходить рейт-лимиты, меняя «IP» в каждом
// запросе. Доверяем только X-Real-IP, который выставляет НАШ доверенный edge-прокси
// (Caddy) и перезаписывает на каждом запросе. Если запрос пришёл напрямую (без edge),
// X-Real-IP не будет — используем RemoteAddr фактического соединения.
//
// Важно для прода: edge должен ОБЯЗАТЕЛЬНО устанавливать X-Real-IP (header_up), а
// прямой доступ к порту приложения, минуя edge, должен быть закрыт на сетевом уровне.
func clientIP(r *http.Request) string {
	if rip := strings.TrimSpace(r.Header.Get("X-Real-IP")); rip != "" {
		return rip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
```

## `internal/usecase/auth.go`

```go
// Аутентификация: регистрация, вход/выход, сессии, восстановление пароля и защита
// рейт-лимитами. Слой usecase — порты (интерфейсы хранилищ и адаптеров) + оркестрация.
// bcrypt живёт здесь: хеширование пароля — деталь алгоритма аутентификации, а не
// хранилища и не транспорта; domain при этом остаётся чистым (валидирует плейн-пароль).
package usecase

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/chudno/zerovibe/internal/domain"
)

// --- порты ---

// UserRepository — порт хранилища пользователей.
type UserRepository interface {
	Create(ctx context.Context, u domain.User) (domain.User, error) // ErrEmailTaken при дубле email
	ByEmail(ctx context.Context, email string) (domain.User, error) // ErrNotFound если нет
	ByID(ctx context.Context, id int64) (domain.User, error)
	UpdatePasswordHash(ctx context.Context, userID int64, hash string) error
	MarkEmailVerified(ctx context.Context, userID int64, at time.Time) error
	CountAdmins(ctx context.Context) (int, error) // для сида первого админа
}

// EmailVerificationRepository — порт хранилища токенов подтверждения почты.
type EmailVerificationRepository interface {
	Create(ctx context.Context, v domain.EmailVerification) error
	ByToken(ctx context.Context, token string) (domain.EmailVerification, error)
	MarkUsed(ctx context.Context, token string, usedAt time.Time) error
	DeleteByUser(ctx context.Context, userID int64) error
}

// SessionRepository — порт хранилища сессий.
type SessionRepository interface {
	Create(ctx context.Context, s domain.Session) error
	ByToken(ctx context.Context, token string) (domain.Session, error) // ErrNotFound если нет
	Delete(ctx context.Context, token string) error
	DeleteByUser(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

// ResetRepository — порт хранилища токенов сброса пароля.
type ResetRepository interface {
	Create(ctx context.Context, p domain.PasswordReset) error
	ByToken(ctx context.Context, token string) (domain.PasswordReset, error)
	MarkUsed(ctx context.Context, token string, usedAt time.Time) error
	DeleteByUser(ctx context.Context, userID int64) error
}

// Email — письмо для отправки через Mailer.
type Email struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Mailer — порт отправки писем (реализация в adapter/platformmail; в тестах — фейк).
type Mailer interface {
	Send(ctx context.Context, m Email) error
}

// RateLimiter — порт оконных счётчиков. Allow атомарно учитывает попытку по ключу и
// сообщает, не превышен ли лимит (и сколько ждать до сброса окна).
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration, now time.Time) (allowed bool, retryAfter time.Duration, err error)
}

// PasswordHasher — порт хеширования пароля (чтобы тесты не платили цену bcrypt-cost).
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error // nil если совпало; domain.ErrInvalidCredentials если нет
}

// SignupPolicy — провайдер флага открытой регистрации (реализуется SettingsService).
// Через него Register узнаёт, разрешена ли регистрация прямо сейчас (настройка
// меняется в рантайме админом).
type SignupPolicy interface {
	Bool(ctx context.Context, key string) (bool, error)
}

// --- bcrypt-реализация PasswordHasher ---

type bcryptHasher struct{ cost int }

// NewBcryptHasher — продакшн-реализация PasswordHasher.
func NewBcryptHasher() PasswordHasher { return bcryptHasher{cost: bcrypt.DefaultCost} }

func (b bcryptHasher) Hash(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), b.cost)
	return string(h), err
}

func (b bcryptHasher) Compare(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return domain.ErrInvalidCredentials
	}
	return err
}

// dummyHash — валидный bcrypt-хеш заглушки. Используется для выравнивания времени
// ответа при несуществующем пользователе (иначе быстрый отказ выдаёт отсутствие email).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("zerovibe-timing-equalizer"), bcrypt.DefaultCost)

// --- конфиг и сервис ---

// RateRule — правило рейт-лимита: не более Limit попыток за Window.
type RateRule struct {
	Limit  int
	Window time.Duration
}

// AuthConfig — настройки бизнес-правил аутентификации.
type AuthConfig struct {
	SessionTTL      time.Duration
	ResetTTL        time.Duration
	VerifyTTL       time.Duration // срок жизни токена подтверждения почты
	AppBaseURL      string        // для ссылки в письме сброса (без хвостового /)
	LoginRateLimit  RateRule
	ForgotRateLimit RateRule
	// Повторная отправка письма подтверждения: два окна (короткое — пауза между
	// письмами, длинное — часовой потолок), оба должны разрешать.
	ResendShortRate RateRule
	ResendHourRate  RateRule
	// SetupToken — код первичной настройки для создания первого админа через /setup.
	// Задаётся снаружи (env SETUP_TOKEN, передаётся плагином при деплое). Пустой →
	// /setup недоступен. Работает только пока в системе нет ни одного админа.
	SetupToken string
}

// AuthService оркеструет регистрацию/вход/сессии/сброс пароля/подтверждение почты.
type AuthService struct {
	users         UserRepository
	sessions      SessionRepository
	resets        ResetRepository
	verifications EmailVerificationRepository
	rl            RateLimiter
	hasher        PasswordHasher
	mailer        Mailer
	settings      SignupPolicy
	cfg           AuthConfig
	now           func() time.Time // подменяется в тестах
}

// NewAuthService собирает сервис аутентификации.
func NewAuthService(
	users UserRepository, sessions SessionRepository, resets ResetRepository,
	verifications EmailVerificationRepository, rl RateLimiter, hasher PasswordHasher,
	mailer Mailer, settings SignupPolicy, cfg AuthConfig,
) *AuthService {
	return &AuthService{
		users: users, sessions: sessions, resets: resets, verifications: verifications,
		rl: rl, hasher: hasher, mailer: mailer, settings: settings, cfg: cfg, now: time.Now,
	}
}

// Ключи настроек, влияющих на аутентификацию.
const (
	signupAllowKey   = "allow_signup"
	requireVerifyKey = "require_email_verification"
)

// Register создаёт обычного пользователя. Если регистрация закрыта → ErrSignupClosed.
// Если включено подтверждение почты — выпускает токен и шлёт письмо со ссылкой.
func (s *AuthService) Register(ctx context.Context, email, password string) (domain.User, error) {
	allowed, err := s.settings.Bool(ctx, signupAllowKey)
	if err != nil {
		return domain.User{}, err
	}
	if !allowed {
		return domain.User{}, domain.ErrSignupClosed
	}
	u, err := domain.NewUser(email, password, domain.RoleUser)
	if err != nil {
		return domain.User{}, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	u.PasswordHash = hash
	created, err := s.users.Create(ctx, u)
	if err != nil {
		// Дубль email → ErrEmailTaken пробрасывается наружу как есть: форма честно
		// сообщает «email уже занят». Это раскрывает наличие аккаунта (user
		// enumeration), но для приложения такого класса принято осознанно: регистрация
		// по умолчанию закрыта (перебирать некому), а при открытой публичной
		// регистрации привычное «email занят» важнее сокрытия факта (как на большинстве
		// сайтов). Вход/сброс/повторное письмо при этом анти-enumeration — там скрытие
		// действительно критично.
		return domain.User{}, err
	}

	// Если требуется подтверждение почты — выпускаем токен и шлём письмо.
	if require, _ := s.settings.Bool(ctx, requireVerifyKey); require {
		_ = s.issueAndSendVerification(ctx, created)
	}
	return created, nil
}

// Login проверяет креды и создаёт сессию. rateKey формирует транспорт (email+IP).
func (s *AuthService) Login(ctx context.Context, email, password, rateKey string) (domain.Session, error) {
	now := s.now()
	allowed, retry, err := s.rl.Allow(ctx, "login:"+rateKey, s.cfg.LoginRateLimit.Limit, s.cfg.LoginRateLimit.Window, now)
	if err != nil {
		return domain.Session{}, err
	}
	if !allowed {
		return domain.Session{}, domain.ErrRateLimited{RetryAfter: retry}
	}

	u, err := s.users.ByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		var nf domain.ErrNotFound
		if errors.As(err, &nf) {
			// Выравниваем время: всё равно прогоняем bcrypt против заглушки, затем
			// отдаём единую ошибку — нельзя отличить «нет email» от «неверный пароль».
			_ = s.hasher.Compare(string(dummyHash), password)
			return domain.Session{}, domain.ErrInvalidCredentials
		}
		return domain.Session{}, err
	}
	if err := s.hasher.Compare(u.PasswordHash, password); err != nil {
		return domain.Session{}, err // ErrInvalidCredentials от hasher
	}

	// Блокируем вход, если требуется подтверждение почты, а она не подтверждена.
	// Проверяем ПОСЛЕ пароля — чтобы не раскрывать статус чужого аккаунта.
	if require, _ := s.settings.Bool(ctx, requireVerifyKey); require && !u.EmailVerified() {
		return domain.Session{}, domain.ErrEmailNotVerified
	}

	token, err := randomToken()
	if err != nil {
		return domain.Session{}, err
	}
	sess := domain.Session{Token: token, UserID: u.ID, CreatedAt: now, ExpiresAt: now.Add(s.cfg.SessionTTL)}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return domain.Session{}, err
	}
	return sess, nil
}

// Authenticate по токену сессии возвращает пользователя (для middleware). Истёкшую
// сессию удаляет и сообщает ErrUnauthenticated.
func (s *AuthService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	sess, err := s.sessions.ByToken(ctx, token)
	if err != nil {
		var nf domain.ErrNotFound
		if errors.As(err, &nf) {
			return domain.User{}, domain.ErrUnauthenticated
		}
		return domain.User{}, err
	}
	if sess.Expired(s.now()) {
		_ = s.sessions.Delete(ctx, token)
		return domain.User{}, domain.ErrUnauthenticated
	}
	return s.users.ByID(ctx, sess.UserID)
}

// Logout удаляет сессию (идемпотентно: отсутствие токена не ошибка).
func (s *AuthService) Logout(ctx context.Context, token string) error {
	err := s.sessions.Delete(ctx, token)
	var nf domain.ErrNotFound
	if err != nil && errors.As(err, &nf) {
		return nil
	}
	return err
}

// RequestReset инициирует сброс пароля. Анти-enumeration: наружу всегда nil (кроме
// рейт-лимита) — по ответу нельзя узнать, есть ли такой email. Письмо уходит только
// если пользователь существует; ошибку отправки логируем внутри, наружу не отдаём.
func (s *AuthService) RequestReset(ctx context.Context, email, rateKey string) error {
	now := s.now()
	allowed, retry, err := s.rl.Allow(ctx, "forgot:"+rateKey, s.cfg.ForgotRateLimit.Limit, s.cfg.ForgotRateLimit.Window, now)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrRateLimited{RetryAfter: retry}
	}

	u, err := s.users.ByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		var nf domain.ErrNotFound
		if errors.As(err, &nf) {
			return nil // молчим: не раскрываем отсутствие email (рейт-лимит уже потрачен)
		}
		return err
	}

	// Инвалидируем прежние токены сброса этого пользователя.
	if err := s.resets.DeleteByUser(ctx, u.ID); err != nil {
		return err
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	pr := domain.PasswordReset{Token: token, UserID: u.ID, CreatedAt: now, ExpiresAt: now.Add(s.cfg.ResetTTL)}
	if err := s.resets.Create(ctx, pr); err != nil {
		return err
	}

	link := s.cfg.AppBaseURL + "/reset?token=" + token
	msg := Email{
		To:      u.Email,
		Subject: "Восстановление пароля",
		Text:    "Чтобы задать новый пароль, перейдите по ссылке:\n" + link + "\n\nЕсли вы не запрашивали сброс — просто проигнорируйте письмо.",
		HTML:    `<p>Чтобы задать новый пароль, перейдите по ссылке:</p><p><a href="` + link + `">` + link + `</a></p><p>Если вы не запрашивали сброс — просто проигнорируйте письмо.</p>`,
	}
	// Ошибку отправки не пробрасываем наружу (анти-enumeration + не валим поток).
	_ = s.mailer.Send(ctx, msg)
	return nil
}

// ConfirmReset проверяет токен, меняет пароль, гасит токен и разлогинивает пользователя
// везде (безопасность: после сброса все старые сессии недействительны).
func (s *AuthService) ConfirmReset(ctx context.Context, token, newPassword string) error {
	if err := domain.ValidatePasswordPlain(newPassword); err != nil {
		return err
	}
	pr, err := s.resets.ByToken(ctx, token)
	if err != nil {
		var nf domain.ErrNotFound
		if errors.As(err, &nf) {
			return domain.ErrInvalidToken
		}
		return err
	}
	if !pr.Usable(s.now()) {
		return domain.ErrInvalidToken
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePasswordHash(ctx, pr.UserID, hash); err != nil {
		return err
	}
	if err := s.resets.MarkUsed(ctx, token, s.now()); err != nil {
		return err
	}
	return s.sessions.DeleteByUser(ctx, pr.UserID)
}

// SetupNeeded сообщает, доступна ли первичная настройка: задан код SetupToken И в
// системе ещё нет ни одного админа. Используется composition root (вывести подсказку
// в лог) и при необходимости транспортом.
func (s *AuthService) SetupNeeded(ctx context.Context) (bool, error) {
	if s.cfg.SetupToken == "" {
		return false, nil
	}
	n, err := s.users.CountAdmins(ctx)
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// Setup создаёт ПЕРВОГО администратора по коду первичной настройки (SetupToken,
// заданному снаружи). Код не задан или админ уже есть → ErrSetupClosed. Неверный
// код → ErrSetupToken. После создания админа CountAdmins>0 → /setup закрыт навсегда.
func (s *AuthService) Setup(ctx context.Context, email, password, token string) error {
	if s.cfg.SetupToken == "" {
		return domain.ErrSetupClosed
	}
	n, err := s.users.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return domain.ErrSetupClosed // админ уже есть — настройка завершена
	}
	// Сравнение в постоянном времени, чтобы код нельзя было подобрать по таймингу.
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.SetupToken)) != 1 {
		return domain.ErrSetupToken
	}

	u, err := domain.NewUser(email, password, domain.RoleAdmin)
	if err != nil {
		return err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u.PasswordHash = hash
	_, err = s.users.Create(ctx, u)
	return err
}

// issueAndSendVerification выпускает токен подтверждения почты (инвалидируя
// прежние) и отправляет письмо со ссылкой. Ошибку отправки наружу не пробрасываем.
func (s *AuthService) issueAndSendVerification(ctx context.Context, u domain.User) error {
	if err := s.verifications.DeleteByUser(ctx, u.ID); err != nil {
		return err
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	now := s.now()
	v := domain.EmailVerification{Token: token, UserID: u.ID, CreatedAt: now, ExpiresAt: now.Add(s.cfg.VerifyTTL)}
	if err := s.verifications.Create(ctx, v); err != nil {
		return err
	}
	link := s.cfg.AppBaseURL + "/verify-email?token=" + token
	msg := Email{
		To:      u.Email,
		Subject: "Подтверждение адреса почты",
		Text:    "Подтвердите адрес почты, перейдя по ссылке:\n" + link + "\n\nЕсли вы не регистрировались — проигнорируйте письмо.",
		HTML:    `<p>Подтвердите адрес почты, перейдя по ссылке:</p><p><a href="` + link + `">` + link + `</a></p><p>Если вы не регистрировались — проигнорируйте письмо.</p>`,
	}
	_ = s.mailer.Send(ctx, msg)
	return nil
}

// ConfirmEmailVerification подтверждает почту по токену: помечает пользователя
// подтверждённым и гасит токен. Невалидный токен → ErrInvalidToken.
func (s *AuthService) ConfirmEmailVerification(ctx context.Context, token string) error {
	v, err := s.verifications.ByToken(ctx, token)
	if err != nil {
		var nf domain.ErrNotFound
		if errors.As(err, &nf) {
			return domain.ErrInvalidToken
		}
		return err
	}
	if !v.Usable(s.now()) {
		return domain.ErrInvalidToken
	}
	now := s.now()
	if err := s.users.MarkEmailVerified(ctx, v.UserID, now); err != nil {
		return err
	}
	return s.verifications.MarkUsed(ctx, token, now)
}

// ResendVerification повторно отправляет письмо подтверждения. Рейт-лимит — два окна
// (пауза между письмами + часовой потолок). Анти-enumeration по ОТВЕТУ: наружу всегда
// nil (кроме рейт-лимита) — по коду/телу ответа нельзя узнать, есть ли email и
// подтверждён ли он. Тайминг при этом не выровнен (для существующего неподтверждённого
// идёт отправка письма, для прочих — ранний выход): это осознанный tradeoff для
// приложения такого класса, тайминг-анализ требует заметных усилий атакующего, а
// рейт-лимит (1/мин) делает массовый замер дорогим.
func (s *AuthService) ResendVerification(ctx context.Context, email, rateKey string) error {
	now := s.now()
	if ok, retry, err := s.rl.Allow(ctx, "resend-short:"+rateKey, s.cfg.ResendShortRate.Limit, s.cfg.ResendShortRate.Window, now); err != nil {
		return err
	} else if !ok {
		return domain.ErrRateLimited{RetryAfter: retry}
	}
	if ok, retry, err := s.rl.Allow(ctx, "resend-hour:"+rateKey, s.cfg.ResendHourRate.Limit, s.cfg.ResendHourRate.Window, now); err != nil {
		return err
	} else if !ok {
		return domain.ErrRateLimited{RetryAfter: retry}
	}

	u, err := s.users.ByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		var nf domain.ErrNotFound
		if errors.As(err, &nf) {
			return nil // молчим: не раскрываем отсутствие email
		}
		return err
	}
	if u.EmailVerified() {
		return nil // уже подтверждена — письмо не нужно, но и не выдаём этот факт
	}
	return s.issueAndSendVerification(ctx, u)
}

// randomToken — 32 байта crypto/rand в hex (64 символа). Для сессий и токенов сброса.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
```

## `internal/usecase/files.go`

```go
package usecase

import (
	"context"
	"io"
)

// FileStore — порт хранилища файлов приложения (картинки, вложения, аватары…).
// Реализация — adapter/platformfiles: платформа выдаёт presigned-URL на S3, сами
// S3-креды в приложение не попадают (как с почтой). Локально — фолбэк на диск.
//
// Модель использования: при сохранении записи с файлом приложение зовёт Save и
// кладёт возвращённый КЛЮЧ в свою запись (колонка в SQLite). Для показа — зовёт URL
// по этому ключу и подставляет ссылку в шаблон. Сами байты в SQLite не хранятся.
type FileStore interface {
	// Save сохраняет содержимое файла и возвращает его ключ (его хранит приложение).
	// size — размер в байтах (0, если неизвестен; тогда передаётся chunked).
	Save(ctx context.Context, fileName string, content io.Reader, size int64) (key string, err error)
	// URL возвращает ссылку для показа/скачивания файла по ключу. Пустой ключ → "".
	URL(ctx context.Context, key string) (string, error)
}
```

## `internal/usecase/notes.go`

```go
// Package usecase содержит бизнес-логику (оркестрацию) и порты — интерфейсы
// репозиториев. Зависит ТОЛЬКО от domain. Конкретные реализации хранилища
// (sqlite) внедряются снаружи и реализуют эти интерфейсы — инверсия зависимостей.
//
// ОБРАЗЕЦ ДЛЯ ГЕНЕРАЦИИ: на каждую сущность — порт (NoteRepository) + сервис
// (NoteService) с методами-операциями. Сервис валидирует через domain-конструкторы
// и делегирует хранение порту. Здесь нет ни SQL, ни HTTP.
package usecase

import (
	"context"

	"github.com/chudno/zerovibe/internal/domain"
)

// NoteRepository — порт хранилища заметок. Реализуется в repository/sqlite.
// Заметки личные: операции привязаны к владельцу (ownerID).
type NoteRepository interface {
	Create(ctx context.Context, n domain.Note) (domain.Note, error) // n.OwnerID уже задан
	ListByOwner(ctx context.Context, ownerID int64) ([]domain.Note, error)
	Delete(ctx context.Context, id, ownerID int64) error // удаляет только свою; иначе ErrNotFound
}

// NoteService — бизнес-операции над заметками.
type NoteService struct {
	repo NoteRepository
}

// NewNoteService собирает сервис с внедрённым репозиторием.
func NewNoteService(repo NoteRepository) *NoteService {
	return &NoteService{repo: repo}
}

// Create валидирует ввод через доменный конструктор и сохраняет заметку владельца.
func (s *NoteService) Create(ctx context.Context, ownerID int64, title, body string) (domain.Note, error) {
	n, err := domain.NewNote(title, body, "")
	if err != nil {
		return domain.Note{}, err
	}
	n.OwnerID = ownerID
	return s.repo.Create(ctx, n)
}

// List возвращает заметки владельца (новые сверху — порядок задаёт репозиторий).
func (s *NoteService) List(ctx context.Context, ownerID int64) ([]domain.Note, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// Delete удаляет заметку владельца по id (чужую не трогает — ErrNotFound).
func (s *NoteService) Delete(ctx context.Context, id, ownerID int64) error {
	return s.repo.Delete(ctx, id, ownerID)
}
```

## `internal/usecase/settings.go`

```go
// Настройки приложения: бизнес-логика поверх реестра доменных настроек.
// Сервис валидирует значения через domain, хранит их в репозитории и отдаёт
// типизированные значения с дефолтами из реестра. Секреты при перечислении
// маскируются — наружу уходит только признак «задано».
package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/chudno/zerovibe/internal/domain"
)

// SettingRepository — порт хранилища настроек. Реализуется в repository/sqlite.
type SettingRepository interface {
	Get(ctx context.Context, key string) (domain.Setting, error) // ErrNotFound если не задано
	Set(ctx context.Context, s domain.Setting) error
	List(ctx context.Context) ([]domain.Setting, error)
}

// SettingsService — операции над настройками приложения.
type SettingsService struct {
	repo SettingRepository
	now  func() time.Time
}

// NewSettingsService собирает сервис настроек.
func NewSettingsService(repo SettingRepository) *SettingsService {
	return &SettingsService{repo: repo, now: time.Now}
}

// Set валидирует значение по реестру и сохраняет настройку.
func (s *SettingsService) Set(ctx context.Context, key, value string) error {
	norm, err := domain.ValidateSetting(key, value)
	if err != nil {
		return err
	}
	return s.repo.Set(ctx, domain.Setting{Key: key, Value: norm, UpdatedAt: s.now().UTC()})
}

// SettingView — представление настройки для перечисления (API/UI). Для секретов
// Value пустой, а Set сообщает, задано ли значение.
type SettingView struct {
	Key   string
	Kind  domain.SettingKind
	Type  domain.SettingType
	Title string
	Value string // для config — текущее/дефолтное значение; для secret — всегда ""
	Set   bool   // задано ли значение (для секретов — единственный наблюдаемый признак)
}

// All перечисляет все известные настройки с текущими значениями. Секреты маскируются.
func (s *SettingsService) All(ctx context.Context) ([]SettingView, error) {
	stored, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]domain.Setting, len(stored))
	for _, st := range stored {
		byKey[st.Key] = st
	}

	defs := domain.SettingDefs()
	views := make([]SettingView, 0, len(defs))
	for _, d := range defs {
		st, set := byKey[d.Key]
		v := SettingView{Key: d.Key, Kind: d.Kind, Type: d.Type, Title: d.Title, Set: set}
		switch d.Kind {
		case domain.SettingSecret:
			// значение не раскрываем
		default:
			if set {
				v.Value = st.Value
			} else {
				v.Value = d.Default
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// raw возвращает сохранённое значение или дефолт из реестра, если не задано.
func (s *SettingsService) raw(ctx context.Context, key string) (string, error) {
	st, err := s.repo.Get(ctx, key)
	if err == nil {
		return st.Value, nil
	}
	var nf domain.ErrNotFound
	if errors.As(err, &nf) {
		if d, ok := domain.LookupSettingDef(key); ok {
			return d.Default, nil
		}
		return "", nil
	}
	return "", err
}

// Bool возвращает значение bool-настройки (или дефолт из реестра).
func (s *SettingsService) Bool(ctx context.Context, key string) (bool, error) {
	v, err := s.raw(ctx, key)
	if err != nil {
		return false, err
	}
	return domain.BoolValue(v), nil
}

// String возвращает значение строковой настройки (или дефолт из реестра).
func (s *SettingsService) String(ctx context.Context, key string) (string, error) {
	return s.raw(ctx, key)
}
```

