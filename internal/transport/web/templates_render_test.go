package web

import (
	"encoding/json"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Тесты ниже ловят ошибки ШАБЛОНОВ, которые НЕ видит `go build`/`go vet`: вызов
// несуществующей функции (`function "seq" not defined`), битый `{{range}}`,
// обращение к несуществующему полю. Такие ошибки проявляются только при ПАРСЕ и
// РЕНДЕРЕ шаблона в рантайме — раньше они роняли приложение лишь при заходе на
// страницу (превью «не ответило»), хотя build/test были зелёными. Теперь падают
// на `make test`, до публикации.

// parseTemplates парсит набор шаблонов приложения из embed (тот же templatesFS и
// тот же вызов ParseFS, что использует сервер в проде — web.go). Ошибка парсинга
// (неизвестная функция, незакрытый action) возвращается здесь.
func parseTemplates() (*template.Template, error) {
	return template.ParseFS(templatesFS, "templates/*.html")
}

// TestTemplatesParse: все шаблоны парсятся. Ловит `function "..." not defined` и
// синтаксические ошибки — ровно то, что убивало страницу в рантайме.
func TestTemplatesParse(t *testing.T) {
	if _, err := parseTemplates(); err != nil {
		t.Fatalf("шаблоны не парсятся (та же ошибка уронила бы страницу в рантайме): %v", err)
	}
}

// TestScreensRender рендерит КАЖДЫЙ экран из манифеста screens.json на его
// фикстуре — как галерея /_screens и как приложение рендерит страницу (layout +
// content по .Page). Ловит render-ошибки: несуществующее поле/метод, nil в
// шаблоне. Нет фикстуры → рендер на пустых данных (тоже без паники).
func TestScreensRender(t *testing.T) {
	// go test cwd = каталог пакета, поэтому screens.json/фикстуры читаем по пути
	// "templates" относительно него (а не через loadScreens/templatesDir — тот путь
	// от корня репо, для dev-галереи).
	const tdir = "templates"
	b, err := os.ReadFile(filepath.Join(tdir, "screens.json"))
	if err != nil {
		t.Skip("screens.json отсутствует — нечего рендерить")
	}
	var screens []screenEntry
	if json.Unmarshal(b, &screens) != nil || len(screens) == 0 {
		t.Skip("screens.json пуст/битый — нечего рендерить")
	}
	tmpl, err := parseTemplates()
	if err != nil {
		t.Fatalf("парс шаблонов: %v", err)
	}
	for _, sc := range screens {
		sc := sc
		t.Run(sc.Name, func(t *testing.T) {
			var pd pageData
			if fb, ferr := os.ReadFile(filepath.Join(tdir, sc.Name+".fixture.json")); ferr == nil {
				_ = json.Unmarshal(fb, &pd)
			}
			pd.Page = sc.Name
			pd.AppName = appName
			// Клон — ExecuteTemplate на общем наборе безопасен для чтения, но клон
			// изолирует состояние между экранами.
			ct, cerr := tmpl.Clone()
			if cerr != nil {
				t.Fatalf("clone: %v", cerr)
			}
			if err := ct.ExecuteTemplate(io.Discard, "layout", pd); err != nil {
				t.Fatalf("экран %q не рендерится (в рантайме страница упала бы): %v", sc.Name, err)
			}
		})
	}
}
