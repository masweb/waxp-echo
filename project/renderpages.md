Flujo de renderizado de páginas

### 1. Servicio público (`ServePage`) — `public.go:29`

Petición entrante → resuelve **site**, **locale** y **slug** desde la URL:

```
GET /nosotros       → locale=default, slug="nosotros"
GET /en/about       → locale="en",    slug="about"
GET /               → locale=default, slug=""
```

Busca en BD la página publicada que coincide con ese slug+locale (`GetPublishedPageSlug`), y luego lee el HTML pre-renderizado de la tabla `page_renders` (`GetPageRenderByPageAndLocale`). Si existe, lo sirve directamente como `text/html`.

### 2. Pre-renderizado (`renderPageForLocale`) — `public.go:163`

Se ejecuta bajo demanda o vía `RegenerateAllPages`. El flujo es:

1. **Obtiene la página** de BD (layout raw con locale maps)
2. **`i18n.Resolve(layout, locale)`** — resuelve todos los `locales` a strings y `menu` a su array por idioma
3. **`i18n.Resolve(options, locale)`** — lo mismo con las opciones del site (header/footer)
4. **Busca SEO** para el locale actual
5. **Llama a `render.Render()`** que produce el HTML completo
6. **`UpsertPageRender`** — guarda el HTML en la tabla `page_renders` (cache por page+locale)

### 3. Motor de renderizado (`internal/render/`)

**`render.go`** — Orquestador:
- Parsea el JSON resuelto a tipos Go (`ParseLayout`, `ParseOptions`)
- Compone las secciones: **header** (si existe) + **layout sections** + **footer** (si existe)
- Llama a `buildHTML(sections, opts, data)`

**`css.go`** — Genera todo el CSS:
- Variables CSS de tema light/dark (`--waxp-text`, `--waxp-bg`, etc.)
- CSS base para bloques (`.b`, `.b-inner`, `.b-overlay`, etc.)
- CSS de tipografía por heading (H1-H6) desde `options.headers`
- Por cada sección: grid CSS, backgrounds, borders, padding
- Por cada bloque: posición en grid (`grid-column`/`grid-row`), font-size fluido (desktop/tablet/mobile con `clamp`)
- 3 media queries: desktop (default), tablet, mobile
- Soporte de `hideOn` por breakpoint
- Font fluido con fórmula `calc(base + slope * vw)` con zoom configurable

**`html.go`** — Genera el HTML:

Estructura del documento:
```
<!DOCTYPE html>
<html lang="{locale}" data-theme="light">
<head>
  <meta charset/viewport>
  <title> / <meta description> / <og:tags>
  <link rel="alternate" hreflang="es" href="...">  (alternates SEO)
  <link canonical>
  <link Google Fonts>
  <style>{todo el CSS inline}</style>
  <script>dark mode init from localStorage</script>
</head>
<body>
  <div class="waxp">
    {header section}
    {layout sections}
    {footer section}
  </div>
  <script>waxpToggleTheme()</script>  (si hay bloque DarkMode)
</body>
</html>
```

Tipos de bloque renderizados (`writeBlockHTML` → switch por tipo):

| Tipo | Función | Qué renderiza |
|---|---|---|
| `Text` | `writeTextBlock` | `<div class="b-tiptap">{HTML de locales.text}</div>` |
| `Image` | `writeImageBlock` | `<img>` con src responsive (desk/tab/mob), fit, link wrapper, alt desde `locales.alt` |
| `Button` | `writeButtonBlock` | `<button>` o `<a>` con colores light/dark, hover/active/focus CSS vars, border radius |
| `Space` | `writeSpaceBlock` | Divider horizontal opcional |
| `DarkMode` | `writeDarkModeBlock` | Botón toggle con iconos SVG luna/sol |
| `LanguageSwitcher` | `writeLangBlock` | `<select>` con opciones por locale, URLs calculadas con `buildRoutePath` |
| `Menu` | `writeMenuBlock` | `<nav><ul>` con colores CSS vars, submenús hover, font custom por nivel |

### 4. Estrategia de caché

El HTML se **pre-renderiza y almacena** en `page_renders` (una fila por `page_id` + `locale_id`). El servicio público lee directamente de ahí — no renderiza en cada petición. Se regenera cuando:

- Se llama a `RegenerateAllPages` (endpoint manual)
- *(Aquí es donde habría que enganchar la regeneración automática al guardar página/site)
