.PHONY: run dev build test vet check css css-watch docker docker-run tidy clean

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

clean:
	rm -rf bin zerovibe.db zerovibe.db-wal zerovibe.db-shm
