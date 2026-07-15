.PHONY: run dev build test vet check css css-watch docker docker-run tidy clean

# Версия standalone-бинаря tailwindcss-extra (Tailwind CSS + DaisyUI внутри,
# Node НЕ нужен). Пин версии — для воспроизводимости сборки.
TW_VERSION := v2.9.1
# TW — какой бинарь звать. Если tailwindcss-extra есть в PATH (в облачном превью он
# вшит в образ раннера) — берём его, НИЧЕГО не качаем. Иначе — локальный ./bin/...
# (правило ниже его скачает). Так шаблон работает и в поде, и на машине разработчика.
TW := $(shell command -v tailwindcss-extra 2>/dev/null || echo ./bin/tailwindcss-extra)
# Ассет под текущую ОС/арх (локально — macOS; в Docker переопределяется).
TW_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/macos/')
TW_ARCH := $(shell uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/')
TW_ASSET := tailwindcss-extra-$(TW_OS)-$(TW_ARCH)
TW_URL := https://github.com/dobicinaitis/tailwind-cli-extra/releases/download/$(TW_VERSION)/$(TW_ASSET)

CSS_IN := assets/input.css
CSS_OUT := internal/transport/web/static/app.css

# Скачать бинарь в ./bin/, если его нет ни в PATH, ни локально. --fail (HTTP-ошибка
# = провал, а не битый файл), --retry (рвущаяся РФ-сеть); chmod +x и проверка
# размера — в ОДНОЙ цели: оборванная закачка не оставит файл без +x, из-за которого
# make css потом падал бы Permission denied. tw-bin — no-op, если TW уже из PATH.
tw-bin:
	@command -v tailwindcss-extra >/dev/null 2>&1 && exit 0; \
	if [ ! -x ./bin/tailwindcss-extra ]; then \
		mkdir -p bin; \
		echo "→ качаю $(TW_ASSET) ($(TW_VERSION))"; \
		curl -sSL --fail --retry 3 --retry-delay 3 -o ./bin/tailwindcss-extra.tmp "$(TW_URL)" \
			&& chmod +x ./bin/tailwindcss-extra.tmp \
			&& mv ./bin/tailwindcss-extra.tmp ./bin/tailwindcss-extra; \
	fi
.PHONY: tw-bin

# Сборка CSS: сканирует html-шаблоны, генерит минимальный CSS (tree-shaking) в
# static/app.css, который вшивается в бинарь через embed. Гоняй после правки
# классов в шаблонах.
css: tw-bin
	$(TW) -i $(CSS_IN) -o $(CSS_OUT) --minify

# Watch-режим для разработки (пересобирает CSS при изменении шаблонов).
css-watch: tw-bin
	$(TW) -i $(CSS_IN) -o $(CSS_OUT) --watch

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
