---
name: daisyui-install
description: Installation notes for daisyUI 5
---

## В ЭТОМ проекте daisyUI УЖЕ установлен и настроен — ставить ничего не нужно

Прежде чем что-либо «устанавливать», знай: в шаблоне daisyUI + Tailwind CSS уже
подключены и работают. **Node/npm НЕ используются.** НЕ выполняй `npm i`, НЕ создавай
`package.json`, `node_modules` или `tailwind.config.js`, НЕ подключай CDN. Официальные
инструкции ниже (npm/CDN) в этом проекте НЕ применяются — оставлены как справка.

Как это устроено здесь:

- CSS собирается standalone-бинарём `tailwindcss-extra` (Tailwind CSS 4 + daisyUI
  внутри самого бинаря, интернет/Node при сборке не нужны). Бинарь скачивается в
  `./bin/` целью Makefile при первой сборке.
- Источник стилей — `assets/input.css`: там `@import "tailwindcss" source(none)`,
  `@plugin "daisyui"` со ВСЕМИ встроенными темами (`light` по умолчанию, `dark` по
  системной настройке; остальные доступны через переключатель тем в шапке) и
  `@source` на каталог шаблонов (tree-shaking — в итоговый CSS попадают только
  реально использованные классы и компоненты).
- Результат — `internal/transport/web/static/app.css`, вшивается в бинарь через `embed`
  и подключён в `layout.html` как `/static/app.css`.
- **После правки классов в шаблонах пересобери CSS: `make css`.** `make run`,
  `make build`, `make check` собирают CSS автоматически перед сборкой/запуском.
- Внешние JS-зависимости отдельных компонентов (напр. библиотека для календаря или
  подсветка кода) в этом проекте недоступны — Node нет. Используй компоненты, которым
  хватает CSS/HTMX; для интерактива — HTMX (см. skill `conventions`).

## daisyUI 5 install notes (официальная справка — в этом проекте НЕ применять)
[install guide](https://daisyui.com/docs/install/)
1. daisyUI 5 requires Tailwind CSS 4
2. `tailwind.config.js` file is deprecated in Tailwind CSS v4. Do not use `tailwind.config.js`. Tailwind CSS v4 only needs `@import "tailwindcss";` in the CSS file if it's a node dependency.
3. daisyUI 5 can be installed using `npm i -D daisyui@latest` and then adding `@plugin "daisyui";` to the CSS file
4. daisyUI is suggested to be installed as a dependency but if you really want to use it from CDN, you can use Tailwind CSS and daisyUI CDN files:
```html
<link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css" />
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
```
5. A CSS file with Tailwind CSS and daisyUI looks like this (if it's a node dependency)
```css
@import "tailwindcss";
@plugin "daisyui";
```
