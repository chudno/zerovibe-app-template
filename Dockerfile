# Multi-stage сборка. Драйвер SQLite — modernc.org/sqlite (чистый Go, без CGO),
# поэтому бинарь статический и кладётся в минимальный distroless-образ без libc.
#
# Образ слушает :8080 и хранит БД в /data (смонтируй volume сюда, чтобы данные
# переживали рестарт контейнера — это и есть persistent-хранилище приложения).

# --- build ---
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache curl ca-certificates

# Кэш зависимостей отдельным слоем.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем CSS ДО go build: standalone-бинарь tailwindcss-extra (Tailwind +
# DaisyUI внутри, Node НЕ нужен) сканирует html-шаблоны и генерит минимальный
# app.css (tree-shaking), который затем вшивается в бинарь через embed. Версия
# пинится для воспроизводимости. Пересборка CSS в образе гарантирует, что вшит
# актуальный стиль, даже если в репо лежит устаревший app.css.
ARG TW_VERSION=v2.9.1
RUN curl -sSL -o /usr/local/bin/tailwindcss-extra \
      https://github.com/dobicinaitis/tailwind-cli-extra/releases/download/${TW_VERSION}/tailwindcss-extra-linux-x64 \
    && chmod +x /usr/local/bin/tailwindcss-extra \
    && tailwindcss-extra -i assets/input.css -o internal/transport/web/static/app.css --minify

# CGO_ENABLED=0 → статический бинарь. -ldflags для уменьшения размера.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/zerovibe ./cmd/server
# Готовим каталог данных в build-стадии: distroless не имеет shell/mkdir/chown,
# поэтому создаём /data здесь и копируем его в runtime с нужным владельцем.
RUN mkdir -p /data

# --- runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/zerovibe /app/zerovibe
# Каталог /data во владении nonroot (uid 65532) — иначе SQLite не создаст файл БД
# в смонтированном volume (ошибка "unable to open database file").
COPY --from=build --chown=nonroot:nonroot /data /data

ENV ADDR=:8080
ENV DB_PATH=file:/data/zerovibe.db
EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["/app/zerovibe"]
