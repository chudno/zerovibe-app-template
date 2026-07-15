#!/usr/bin/env bash
# Генерирует PROJECT_CONTEXT.md — весь код шаблона (путь + содержимое каждого
# файла) в одном файле. Агент читает его ПЕРВЫМ вместо десятков Read/Glob по
# файлам: токены те же (он всё равно всё читает), но без множества тулколов —
# быстрее первый ответ. Пересобирать при изменении структуры: `make context`.
#
# Включаем: исходники (.go без тестов), шаблоны (.html), миграции (.sql),
# go.mod, Makefile, input.css. Исключаем: тесты, сборочный мусор, .git, bin,
# tmp, node_modules, собранный app.css, сам PROJECT_CONTEXT.md, скиллы (.claude).
set -euo pipefail
cd "$(dirname "$0")/.."

OUT="PROJECT_CONTEXT.md"

# Список файлов в детерминированном порядке (совместимо с bash 3.2 — без mapfile).
FILES=$(
  find . -type f \
    \( -name '*.go' -o -name '*.html' -o -name '*.sql' \
       -o -name 'go.mod' -o -name 'Makefile' -o -name 'input.css' \) \
    -not -path './.git/*' \
    -not -path './bin/*' \
    -not -path './tmp/*' \
    -not -path '*/node_modules/*' \
    -not -path './.claude/*' \
    -not -name '*_test.go' \
    -not -path '*/static/app.css' \
  | sed 's|^\./||' | LC_ALL=C sort
)

COUNT=$(printf '%s\n' "$FILES" | grep -c . || true)

{
  echo "# PROJECT_CONTEXT — код шаблона одним файлом"
  echo
  echo "Автоген (\`make context\`). Здесь путь и содержимое каждого файла шаблона —"
  echo "читай ЭТОТ файл, чтобы понять структуру и паттерны, вместо изучения файлов"
  echo "по отдельности. Пересобирается при изменении шаблона; НЕ правь вручную."
  echo
  echo "## Файлы ($COUNT)"
  echo
  printf '%s\n' "$FILES" | while IFS= read -r f; do
    [ -n "$f" ] && echo "- \`$f\`"
  done
  echo
  echo "---"
  echo
  printf '%s\n' "$FILES" | while IFS= read -r f; do
    [ -n "$f" ] || continue
    case "$f" in
      *.go)     lang=go ;;
      *.html)   lang=html ;;
      *.sql)    lang=sql ;;
      *.css)    lang=css ;;
      go.mod)   lang=go ;;
      Makefile) lang=make ;;
      *)        lang="" ;;
    esac
    echo "## \`$f\`"
    echo
    echo "\`\`\`$lang"
    cat "$f"
    echo "\`\`\`"
    echo
  done
} > "$OUT"

echo "→ $OUT собран: $COUNT файлов, $(wc -l < "$OUT" | tr -d ' ') строк, $(wc -c < "$OUT" | tr -d ' ') байт"
