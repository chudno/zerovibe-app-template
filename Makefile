.PHONY: run dev build test vet check docker docker-run tidy clean

# Локальный запуск (БД в текущем каталоге).
run:
	go run ./cmd/server

# Dev-режим live-reload: ZV_DEV=1 заставляет приложение читать html-шаблоны и
# статику С ДИСКА на каждый запрос (а не из вшитого embed). Правки .html видны
# сразу по F5, без пересборки бинаря. Правки .go требуют перезапуска (Ctrl-C +
# `make dev`) — бинарь надо собрать заново.
dev:
	ZV_DEV=1 SECURE_COOKIE=false go run ./cmd/server

# Проверка перед публикацией: сборка + статанализ + тесты.
check:
	go build ./... && go vet ./... && go test ./...

# Сборка бинаря.
build:
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
