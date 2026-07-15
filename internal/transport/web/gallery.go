package web

// Галерея экранов /_screens — служебная страница ТОЛЬКО в dev-режиме (ZV_DEV=1),
// для фазы 1 двухфазной сборки (сначала вёрстка, потом логика). Показывает вид
// экранов-заготовок ДО того, как к ним привязаны роуты/данные: слева список
// экранов из templates/screens.json, справа iframe с рендером выбранного экрана.
//
// Экран рендерится ТЕМ ЖЕ layout, что и рабочее приложение (ExecuteTemplate
// "layout"), но данные берутся из фикстуры templates/{name}.fixture.json, а не из
// БД. Один шаблон — два источника данных (галерея / приложение), НОЛЬ копий → нет
// рассинхрона. Ключи фикстуры = контракт полей, которые фаза 2 заполнит реально.
//
// В проде роуты не монтируются (см. Routes: блок if s.dev), путь недостижим.

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

// templatesDir — каталог шаблонов на диске (dev): там же screens.json и фикстуры.
const templatesDir = "internal/transport/web/templates"

// screenEntry — запись манифеста экранов (templates/screens.json).
type screenEntry struct {
	Name  string `json:"name"`  // идентификатор (= pageData.Page)
	Title string `json:"title"` // человеческое имя для списка
	File  string `json:"file"`  // имя html-файла (для справки)
}

// loadScreens читает манифест экранов. Пусто при ошибке — галерея покажет пустой
// список, но не упадёт.
func loadScreens() []screenEntry {
	b, err := os.ReadFile(filepath.Join(templatesDir, "screens.json"))
	if err != nil {
		return nil
	}
	var s []screenEntry
	if json.Unmarshal(b, &s) != nil {
		return nil
	}
	return s
}

// isKnownScreen — экран есть в манифесте (защита от path traversal: name идёт в
// имя файла фикстуры).
func isKnownScreen(name string) bool {
	for _, s := range loadScreens() {
		if s.Name == name {
			return true
		}
	}
	return false
}

// handleScreenRender рендерит ОДИН экран-заготовку с данными из его фикстуры —
// точно так же, как renderPage рендерит рабочую страницу (layout + content по
// .Page), только данные из {name}.fixture.json. Нет фикстуры / битый JSON → экран
// рендерится без данных (пустые списки, метки шаблона) — не 500.
func (s *Server) handleScreenRender(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !isKnownScreen(name) {
		http.NotFound(w, r)
		return
	}
	// Данные из фикстуры → pageData (НЕ голый map: шаблоны зовут методы time.Time,
	// напр. note.CreatedAt.Format; json.Unmarshal строки RFC3339 в time.Time даёт
	// рабочий метод). Нет файла/битый JSON → рендерим с пустыми данными.
	var pd pageData
	if b, err := os.ReadFile(filepath.Join(templatesDir, name+".fixture.json")); err == nil {
		_ = json.Unmarshal(b, &pd) // ошибку игнорируем: частично заполнит, чего хватило
	}
	pd.Page = name       // на всякий случай — layout выбирает content по .Page
	pd.AppName = appName // галерея не зовёт settings-сервис; имя приложения — константа
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates().ExecuteTemplate(w, "layout", pd); err != nil {
		http.Error(w, "ошибка рендера экрана: "+err.Error(), http.StatusInternalServerError)
	}
}

// galleryShellTmpl — оболочка галереи (список слева + iframe справа). Отдаётся
// отдельным html.Template (НЕ через основной набор шаблонов приложения), чтобы её
// define'ы не попадали в набор экранов. Свой минимальный CSS — хром галереи не
// зависит от темы приложения.
var galleryShellTmpl = template.Must(template.New("gallery").Parse(`<!doctype html>
<html lang="ru"><head><meta charset="utf-8"><title>Экраны</title>
<style>
  *{box-sizing:border-box} html,body{margin:0;height:100%;font-family:system-ui,-apple-system,sans-serif}
  .wrap{display:flex;height:100vh}
  .side{width:240px;flex-shrink:0;border-right:1px solid #e5e7eb;background:#fafafa;overflow-y:auto;padding:12px}
  .side h2{margin:4px 8px 12px;font-size:12px;letter-spacing:.04em;text-transform:uppercase;color:#9ca3af}
  .side a{display:block;padding:8px 10px;margin-bottom:2px;border-radius:8px;color:#374151;text-decoration:none;font-size:14px}
  .side a:hover{background:#eef2f7}
  .side a.active{background:#2b8fe6;color:#fff}
  .main{flex:1;min-width:0}
  iframe{width:100%;height:100%;border:0;background:#fff}
</style></head>
<body><div class="wrap">
  <nav class="side"><h2>Экраны</h2>
    {{range .Screens}}<a href="#{{.Name}}" data-name="{{.Name}}">{{.Title}}</a>{{end}}
  </nav>
  <div class="main"><iframe id="frame" src="/_screens/{{.First}}"></iframe></div>
</div>
<script>
  var links=document.querySelectorAll('.side a'), frame=document.getElementById('frame');
  function show(name){
    frame.src='/_screens/'+name;
    links.forEach(function(a){a.classList.toggle('active',a.dataset.name===name)});
    if(location.hash!=='#'+name) location.hash=name;
  }
  links.forEach(function(a){a.addEventListener('click',function(e){e.preventDefault();show(a.dataset.name)})});
  var initial=(location.hash||'').replace('#','')||{{.First}};
  show(initial);
</script>
</body></html>`))

// handleScreensGallery отдаёт оболочку галереи со списком экранов.
func (s *Server) handleScreensGallery(w http.ResponseWriter, r *http.Request) {
	screens := loadScreens()
	first := ""
	if len(screens) > 0 {
		first = screens[0].Name
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = galleryShellTmpl.Execute(w, map[string]any{"Screens": screens, "First": first})
}
