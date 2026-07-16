// E2E-тесты транспорта на полном стеке (реальный SQLite во временном файле).
package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chudno/zerovibe/internal/platform/db"
	"github.com/chudno/zerovibe/internal/repository/sqlite"
	"github.com/chudno/zerovibe/internal/usecase"
)

// noMailer — заглушка Mailer для тестов транспорта (письма тут не нужны).
type noMailer struct{}

func (noMailer) Send(_ context.Context, _ usecase.Email) error { return nil }

// capturedToken — токен из последнего письма сброса (для e2e reset-флоу).
var capturedToken string

// captureMailer достаёт токен сброса из ссылки в тексте письма.
type captureMailer struct{}

var resetLinkRe = regexp.MustCompile(`/reset\?token=([0-9a-f]+)`)
var verifyLinkRe = regexp.MustCompile(`/verify-email\?token=([0-9a-f]+)`)

// capturedVerifyToken — токен из последнего письма подтверждения почты.
var capturedVerifyToken string

func (captureMailer) Send(_ context.Context, m usecase.Email) error {
	if mm := resetLinkRe.FindStringSubmatch(m.Text); mm != nil {
		capturedToken = mm[1]
	}
	if mm := verifyLinkRe.FindStringSubmatch(m.Text); mm != nil {
		capturedVerifyToken = mm[1]
	}
	return nil
}

// buildStackWithMailer — как buildStack, но с мейлером-перехватчиком токена сброса.
// testSetupToken — код первичной настройки, заданный в тестовом стеке. Совпадает с
// тем, что в проде задаёт плагин через env SETUP_TOKEN. /setup доступен, только пока
// в системе нет ни одного админа.
const testSetupToken = "test-setup-token"

func buildStackWithMailer(t *testing.T, allowSignup bool) (http.Handler, *usecase.AuthService, *usecase.SettingsService) {
	t.Helper()
	capturedToken = ""
	capturedVerifyToken = ""
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("открыть БД: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.MigrateUp(context.Background()); err != nil {
		t.Fatalf("миграции: %v", err)
	}
	settings := usecase.NewSettingsService(sqlite.NewSettingRepo(database))
	// Явно выставляем allow_signup под тест (не полагаемся на дефолт реестра — он
	// теперь true, поэтому «закрытую» регистрацию нужно ставить явно false).
	if allowSignup {
		_ = settings.Set(context.Background(), "allow_signup", "true")
	} else {
		_ = settings.Set(context.Background(), "allow_signup", "false")
	}
	auth := usecase.NewAuthService(
		sqlite.NewUserRepo(database), sqlite.NewSessionRepo(database),
		sqlite.NewResetRepo(database), sqlite.NewEmailVerificationRepo(database), sqlite.NewRateLimitRepo(database),
		usecase.NewBcryptHasher(), captureMailer{}, settings,
		usecase.AuthConfig{
			SessionTTL:      time.Hour,
			ResetTTL:        time.Hour,
			VerifyTTL:       24 * time.Hour,
			AppBaseURL:      "http://localhost",
			LoginRateLimit:  usecase.RateRule{Limit: 5, Window: 15 * time.Minute},
			ForgotRateLimit: usecase.RateRule{Limit: 3, Window: time.Hour},
			ResendShortRate: usecase.RateRule{Limit: 1, Window: time.Minute},
			ResendHourRate:  usecase.RateRule{Limit: 5, Window: time.Hour},
			SetupToken:      testSetupToken,
		},
	)
	srv, err := NewServer(auth, settings, Config{SecureCookie: false, CookieName: "zv_session"})
	if err != nil {
		t.Fatalf("сервер: %v", err)
	}
	return srv.Routes(), auth, settings
}

// buildStack собирает полный стек на временной БД. Возвращает handler и сервисы для
// сидирования. allowSignup управляет открытой регистрацией.
func buildStack(t *testing.T, allowSignup bool) (http.Handler, *usecase.AuthService, *usecase.SettingsService) {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("открыть БД: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.MigrateUp(context.Background()); err != nil {
		t.Fatalf("миграции: %v", err)
	}

	settings := usecase.NewSettingsService(sqlite.NewSettingRepo(database))
	// Явно ставим allow_signup под тест: дефолт реестра теперь true, поэтому
	// «закрытую» регистрацию для тестов надо выставлять явно false.
	val := "false"
	if allowSignup {
		val = "true"
	}
	if err := settings.Set(context.Background(), "allow_signup", val); err != nil {
		t.Fatalf("настройка allow_signup: %v", err)
	}
	auth := usecase.NewAuthService(
		sqlite.NewUserRepo(database), sqlite.NewSessionRepo(database),
		sqlite.NewResetRepo(database), sqlite.NewEmailVerificationRepo(database), sqlite.NewRateLimitRepo(database),
		usecase.NewBcryptHasher(), noMailer{}, settings,
		usecase.AuthConfig{
			SessionTTL:      time.Hour,
			ResetTTL:        time.Hour,
			VerifyTTL:       24 * time.Hour,
			AppBaseURL:      "http://localhost",
			LoginRateLimit:  usecase.RateRule{Limit: 5, Window: 15 * time.Minute},
			ForgotRateLimit: usecase.RateRule{Limit: 3, Window: time.Hour},
			ResendShortRate: usecase.RateRule{Limit: 1, Window: time.Minute},
			ResendHourRate:  usecase.RateRule{Limit: 5, Window: time.Hour},
			SetupToken:      testSetupToken,
		},
	)
	srv, err := NewServer(auth, settings, Config{SecureCookie: false, CookieName: "zv_session"})
	if err != nil {
		t.Fatalf("сервер: %v", err)
	}
	return srv.Routes(), auth, settings
}

// buildStackPreview собирает стек в режиме превью (PreviewMode=true) — для проверки
// атрибутов сессионной cookie в cross-site iframe (SameSite=None; Secure).
func buildStackPreview(t *testing.T) (http.Handler, *usecase.AuthService) {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("открыть БД: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.MigrateUp(context.Background()); err != nil {
		t.Fatalf("миграции: %v", err)
	}
	settings := usecase.NewSettingsService(sqlite.NewSettingRepo(database))
	auth := usecase.NewAuthService(
		sqlite.NewUserRepo(database), sqlite.NewSessionRepo(database),
		sqlite.NewResetRepo(database), sqlite.NewEmailVerificationRepo(database), sqlite.NewRateLimitRepo(database),
		usecase.NewBcryptHasher(), noMailer{}, settings,
		usecase.AuthConfig{
			SessionTTL: time.Hour, ResetTTL: time.Hour, VerifyTTL: 24 * time.Hour,
			AppBaseURL:      "http://localhost",
			LoginRateLimit:  usecase.RateRule{Limit: 5, Window: 15 * time.Minute},
			ForgotRateLimit: usecase.RateRule{Limit: 3, Window: time.Hour},
			ResendShortRate: usecase.RateRule{Limit: 1, Window: time.Minute},
			ResendHourRate:  usecase.RateRule{Limit: 5, Window: time.Hour},
			SetupToken:      testSetupToken,
		},
	)
	srv, err := NewServer(auth, settings, Config{SecureCookie: false, CookieName: "zv_session", PreviewMode: true})
	if err != nil {
		t.Fatalf("сервер: %v", err)
	}
	return srv.Routes(), auth
}

// seedAdminAndLogin создаёт первого админа (через первичную настройку по тестовому
// коду) и возвращает cookie его сессии. Идёт тем же путём, что и прод — /setup.
func seedAdminAndLogin(t *testing.T, h http.Handler, auth *usecase.AuthService, email, pass string) *http.Cookie {
	t.Helper()
	if err := auth.Setup(context.Background(), email, pass, testSetupToken); err != nil {
		t.Fatalf("сид админа: %v", err)
	}
	return loginCookie(t, h, email, pass)
}

// loginCookie логинится и возвращает cookie сессии.
func loginCookie(t *testing.T, h http.Handler, email, pass string) *http.Cookie {
	t.Helper()
	form := url.Values{"email": {email}, "password": {pass}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	for _, c := range rec.Result().Cookies() {
		if c.Name == "zv_session" && c.Value != "" {
			return c
		}
	}
	t.Fatalf("логин не вернул cookie сессии (код %d): %s", rec.Code, rec.Body.String())
	return nil
}

// newAuthedServer — стек + cookie залогиненного пользователя (через сид-админа).
func newAuthedServer(t *testing.T) (http.Handler, *http.Cookie) {
	h, auth, _ := buildStack(t, false)
	c := seedAdminAndLogin(t, h, auth, "owner@example.com", "password123")
	return h, c
}

func TestStaticServed(t *testing.T) {
	h, _ := newAuthedServer(t)
	// app.css больше нет (DaisyUI убрали) — проверяем оставшуюся статику: htmx и шрифт.
	for _, name := range []string{"htmx.min.js", "fonts/onest-400.woff2"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/static/"+name, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("GET /static/%s: ожидался 200, получен %d", name, rec.Code)
		}
		if rec.Body.Len() == 0 {
			t.Errorf("GET /static/%s: пустое тело", name)
		}
	}
}
