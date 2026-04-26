# Render Pages

El sistema de renderizado genera HTML estático completo por cada página y locale, lo almacena en BD y lo sirve directamente en las peticiones públicas sin procesamiento en tiempo real.

---

## Arquitectura general

```
Petición pública (GET /nosotros)
  │
  ▼
ServePage() ──► Lee page_renders (HTML cacheado)
  │                      │
  │                      ├─ Hit → devuelve HTML directamente
  │                      └─ Miss → 404 "page not rendered yet"
  │
  ▼
Regeneración (manual o automática)
  │
  ▼
renderPageForLocale(pageID, locale)
  │
  ├─ 1. GetPageByID          → layout raw (con locale maps)
  ├─ 2. i18n.Resolve(layout) → layout resuelto a strings
  ├─ 3. i18n.Resolve(options) → opciones resueltas (header/footer)
  ├─ 4. GetPageSeoByPageID   → SEO para el locale
  ├─ 5. render.Render()      → genera HTML completo
  └─ 6. UpsertPageRender     → guarda en page_renders
```

---

## Tabla `page_renders`

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | int64 | ID del registro |
| `page_id` | int64 | FK a la página |
| `locale_id` | int64 | FK al locale |
| `html` | string | HTML renderizado completo |
| `updated_at` | timestamptz | Fecha de última renderización |

Clave única: `(page_id, locale_id)`. El `UpsertPageRender` hace `INSERT ... ON CONFLICT UPDATE`.

---

## Flujo del servicio público (`ServePage`)

**Archivo:** `internal/handler/public.go`

### Resolución de URL

```
GET /                → locale = default, slug = ""
GET /nosotros        → locale = default, slug = "nosotros"
GET /en/about        → locale = "en",     slug = "about"
GET /en              → locale = "en",     slug = ""
```

El handler:

1. Obtiene el site activo (`GetLiveSite`)
2. Lista los locales del site
3. Parsea la URL para extraer `localeCode` y `slug`:
   - Si el primer segmento es un locale conocido → usa ese locale
   - Si no → es el locale por defecto, el segmento es parte del slug
4. Busca la página publicada con ese slug+locale (`GetPublishedPageSlug`)
5. Lee el HTML pre-renderizado de `page_renders`
6. Devuelve `text/html; charset=utf-8`

### Resolución de rutas (`buildRoutePath`)

| Locale | isDefault | Slug | Ruta |
|--------|-----------|------|------|
| `es` | `true` | `""` | `/` |
| `en` | `false` | `""` | `/en` |
| `es` | `true` | `"nosotros"` | `/nosotros` |
| `en` | `false` | `"about"` | `/en/about` |

---

## Regeneración de páginas

### Manual: `RegenerateAllPages`

**Archivo:** `internal/handler/public.go`

```
POST /api/sites/:id/regenerate
```

Recorre todas las páginas publicadas × todos los locales y llama a `renderPageForLocale` para cada combinación. Devuelve:

```json
{ "regenerated": 12 }
```

### Automática: `SetLive`

**Archivo:** `internal/handler/site.go`

```
PUT /api/sites/:id/live
```

Al marcar un site como live:

1. Desactiva el site que estuviera live (`ClearLiveSites`)
2. Activa el nuevo site (`ActivateSiteLive`)
3. **Regenera automáticamente** todas las páginas publicadas del site para todos los locales

---

## Función `renderPageForLocale`

**Archivo:** `internal/handler/public.go`

Pipeline de renderizado:

1. **Obtiene la página** de BD → layout con locale maps (`{"es": "Hola", "en": "Hello"}`)
2. **`i18n.Resolve(layout, locale)`** → resuelve todos los `locales` a strings, y `menu` de `{locale: [...]}` a `[...]`
3. **`i18n.Resolve(options, locale)`** → lo mismo con header/footer del site
4. **Busca SEO** para el locale actual → título y descripción
5. **Busca slugs** de la página para generar `<link hreflang>`
6. **`render.Render()`** → genera el HTML completo
7. **`UpsertPageRender`** → guarda en `page_renders`

---

## Motor de renderizado (`internal/render/`)

### Tipos (`types.go`)

Estructuras Go tipadas que mapean el JSON resuelto:

- **`PageData`** — datos completos para renderizar: layout, options, SEO, locale, slugs, domain
- **`Section`** — sección con ID, bloques, breakpoints (mobile/tablet/desktop) y estilo
- **`Block`** — bloque con tipo, coordenadas por breakpoint, estilo, y campos específicos por tipo
- **`SiteOptions`** — configuración global (ver sección dedicada)
- **`BlockIcon`** — `{name, strokeWidth}` para el bloque Icon

### SiteOptions

| Campo | Tipo | Descripción | Default |
|-------|------|-------------|---------|
| `lightColor` | string | Color texto light | — |
| `darkColor` | string | Color texto dark | — |
| `lightBackColor` | string | Color fondo light | — |
| `darkBackColor` | string | Color fondo dark | — |
| `lightAccentColor` | string | Color acento links light | fallback a `lightColor` |
| `darkAccentColor` | string | Color acento links dark | fallback a `darkColor` |
| `fontSize` | float64 | Font-size base (rem) | — |
| `lineHeight` | float64 | Line-height base (em) | — |
| `desktopWidth` | int | Ancho máximo (px) | — |
| `desktopTextZoom` | float64 | Zoom texto desktop | 1 |
| `desktopMargin` | float64 | Padding lateral desktop (px) | 0 |
| `tabletBP` | int | Breakpoint tablet (px) | 1024 |
| `tabletTextZoom` | float64 | Zoom texto tablet | 1 |
| `tabletMargin` | float64 | Padding lateral tablet (px) | 0 |
| `mobileBP` | int | Breakpoint mobile (px) | 767 |
| `mobileTextZoom` | float64 | Zoom texto mobile | 2.6 |
| `mobileMargin` | float64 | Padding lateral mobile (px) | 0 |
| `globalFontFamily` | Font | `{family, weight, italic}` | — |
| `fonts` | []SiteFont | Fuentes Google Fonts | — |
| `headers` | json | Config tipografía H1-H6 | — |
| `header` | *Section | Sección header fija | — |
| `footer` | *Section | Sección footer | — |
| `darkMode` | bool | Estado dark mode | — |

### Coordenadas por breakpoint

Cada bloque tiene 3 sets de coordenadas:

| Campo | Descripción |
|-------|-------------|
| `d` (desktop) | `{x, y, w, h}` posición en el grid desktop |
| `t` (tablet) | `{x, y, w, h}` posición en el grid tablet |
| `m` (mobile) | `{x, y, w, h}` posición en el grid mobile |

### Orquestador (`render.go`)

```
Render(input) → ParseLayout + ParseOptions → RenderPage(data)
                                         │
                                         ▼
                              header + sections + footer
                                         │
                                         ▼
                              buildHTML(sections, opts, data)
                                         │
                                    buildCSS + writeSectionHTML × N
```

Compone las secciones en orden:

1. **Header** (si existe en `options.header`) — con `isFixed: true`, prefix `sh`
2. **Sections del layout** — prefix `s{id}`
3. **Footer** (si existe en `options.footer`) — con `isFixed: true`, prefix `sf`

### Generación de CSS (`css.go`)

El CSS se genera inline dentro de `<style>` y contiene:

**Base:**
- Reset básico (`* { margin: 0; padding: 0; box-sizing: border-box }`)
- Estilos de `body`: font-smoothing, text-rendering, text-size-adjust
- Scrollbar custom: `::-webkit-scrollbar { width: 4px; height: 4px }`
- Variables CSS de tema: `--waxp-text`, `--waxp-bg`, `--waxp-accent`, `--waxp-ff`, `--waxp-fs`, `--waxp-lh`, `--waxp-dw`
- Variables dark: `:root[data-theme="dark"] { --waxp-text; --waxp-bg; --waxp-accent }`
- Padding lateral en `.waxp` según márgenes configurados (desktop/tablet/mobile)
- Estilos base de bloques: `.b`, `.b-inner`, `.b-overlay`, `.b-center`, `.b-link`
- Todos los elementos interactivos (`button`, `select`) usan `font:inherit` para heredar font-size/family fluidos
- Estilos de texto (tiptap): párrafos, imágenes, enlaces (color: `--waxp-accent`), listas, blockquotes
- Estilos de componentes: botones, dark mode toggle, language switcher, iconos, menús

**Por sección:**
- Grid CSS: `grid-template-columns`, `grid-template-rows`, `gap`
- `aspect-ratio` si hay rows definidas
- `max-width` (desde `options.desktopWidth` o `section.style.maxWidth`)
- Backgrounds (color/gradiente/imagen) para light y dark
- Padding y margin
- `position: sticky` si es header/footer

**Por bloque:**
- Posición en grid: `grid-column: X/span W; grid-row: Y/span H`
- Background, border, padding
- `overflow: visible` para bloques tipo Menu (submenús)
- Font-size fluido (ver sección de tipografía)
- Color light/dark (`blk.Color`/`blk.DarkColor`) vía `<style>` scoped — nunca inline
- `hideOn` por breakpoint → `display: none !important`

**Media queries:**

| Breakpoint | Default | Configurable en |
|------------|---------|-----------------|
| Tablet | `≤1024px` | `options.tabletBP` |
| Mobile | `≤767px` | `options.mobileBP` |

Cada media query redefine grid, coordenadas de bloques, font-size y padding de `.waxp`.

En mobile, el `<select>` del LanguageSwitcher se resetea a `appearance:auto` con `font-size:16px!important` para que el dropdown nativo funcione correctamente.

**Tipografía fluida:**

Desktop: `font-size: calc(base_px + slope * vw)` con clamp en `targetWidth`. Si hay `desktopMargin`, se aplica corrección: `calc(base_px + slope * vw - correction_px)`.

Tablet/Mobile: `font-size: calc((fontSize + zoom)vw - correction_px)` donde zoom viene de `options.tabletTextZoom` / `options.mobileTextZoom`. La corrección ajusta por el padding lateral (`effectiveWidth = rawWidth - 2 * margin`).

**Headers (H1-H6):**

`options.headers` es un JSON con configuración por tag:
```json
{
  "h1": { "size": 2.5, "family": "Merriweather", "weight": 700, "lineHeight": 1.2 },
  "h2": { "size": 2.0, "family": "Merriweather", "weight": 600 }
}
```

Se genera CSS específico: `.b-tiptap h1 { font-size: 2.5em; ... }`.

**Google Fonts:**

Se genera una URL `https://fonts.googleapis.com/css2?family=...` con todas las fuentes declaradas en `options.fonts`, incluyendo weights e italics.

### Generación de HTML (`html.go`)

Estructura del documento:

```html
<!DOCTYPE html>
<html lang="{locale}" data-theme="light">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  
  <!-- SEO -->
  <title>{seo.title}</title>
  <meta name="description" content="{seo.description}">
  <meta property="og:title" content="{seo.title}">
  <meta property="og:description" content="{seo.description}">
  <meta property="og:type" content="website">
  
  <!-- hreflang (una por slug locale) -->
  <link rel="alternate" hreflang="es" href="https://domain.com/nosotros">
  <link rel="alternate" hreflang="en" href="https://domain.com/en/about">
  
  <!-- Canonical -->
  <link rel="canonical" href="https://domain.com/nosotros">
  
  <!-- Google Fonts -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=..." rel="stylesheet">
  
  <!-- CSS inline -->
  <style>{todo el CSS generado}</style>
  
  <!-- Dark mode init -->
  <script>(localStorage/mediaQuery init)</script>
</head>
<body>
  <div class="waxp">
    <!-- Header section -->
    <div class="sh-row sr">
      <div class="sh s">
        {bloques del header}
      </div>
    </div>
    
    <!-- Layout sections -->
    <div class="s1-row sr">
      <div class="s1 s">
        {bloques de la sección 1}
      </div>
    </div>
    
    <!-- Footer section -->
    <div class="sf-row sr">
      <div class="sf s">
        {bloques del footer}
      </div>
    </div>
  </div>
  
  <!-- Dark mode toggle script (si hay bloque DarkMode) -->
  <script>function waxpToggleTheme() {...}</script>
</body>
</html>
```

**Estructura de cada bloque:**

```html
<div class="{prefix}-b{id} b">
  <!-- Overlay de background (si aplica) -->
  <div class="b-overlay" style="..."></div>
  
  <div class="b-inner [b-center]">
    <!-- Contenido específico por tipo -->
  </div>
</div>
```

Los bloques `DarkMode`, `LanguageSwitcher`, `Button` e `Icon` reciben clase `b-center` para centrado vertical y horizontal.

**Backgrounds con overlay:**

Los backgrounds de imagen se renderizan como overlays absolutos (`position: absolute; inset: 0`) con `pointer-events: none`. Si el tema dark tiene un overlay diferente, se inyecta un `<style>` scoped:

```html
<div class="b-overlay" style="background-image: url('...')"></div>
<style>:root[data-theme="dark"] .s1-b5 .b-overlay { background-image: url('...') }</style>
```

**Color de bloques:**

Todos los colores theme-dependientes (bg, text, border) van en `<style>` con selectores light/dark — nunca inline. Esto evita que los estilos inline impidan la alternancia de tema.

Light: `.{selector} .btn-b { background-color: ...; color: ...; }`
Dark: `:root[data-theme="dark"] .{selector} .btn-b { background-color: ...; color: ...; }`

---

## Tipos de bloque

### Text (`writeTextBlock`)

```html
<div class="b-tiptap">{locales.text}</div>
```

El contenido de `locales.text` es HTML raw generado por TipTap. Los enlaces usan `color: var(--waxp-accent)`.

### Image (`writeImageBlock`)

```html
<a class="b-link" href="{link.url}" target="_blank" rel="noopener noreferrer">
  <img src="{mediaBase}{image.url_desk}" alt="{locales.alt}" 
       style="width:100%;height:auto;" loading="lazy">
</a>
```

- Soporta 3 fuentes: `url_desk`, `url_tab`, `url_mob` (actualmente usa `url_desk`)
- `fit`: `cover`, `height`, o default (`width:100%`)
- Link wrapper si `block.link` existe (internal o external)

### Button (`writeButtonBlock`)

```html
<div class="btn-wrap">
  <a class="btn-b" href="..." style="width:60%;padding:12px 24px;border-radius:8px;--btn-hover-bg:...">
    {locales.label}
  </a>
  <style>
    .{selector} .btn-b { background-color:...; color:...; --btn-hover-bg:...; }
    .{selector} .btn-b:hover { background-color: var(--btn-hover-bg)!important; }
    .{selector} .btn-b:active { background-color: var(--btn-active-bg)!important; }
    .{selector} .btn-b:focus { box-shadow: var(--btn-focus-ring); outline: none; }
    :root[data-theme="dark"] .{selector} .btn-b { background-color:...; color:...; }
  </style>
</div>
```

- `font:inherit` hereda font-size fluido y font-family del site
- Colores light/dark para: bg, text, border, hover, active, focus — todo en `<style>`, nunca inline
- CSS custom properties para estados interactivos (`--btn-hover-bg`, etc.)
- Fallback de color de texto a `blk.Color`/`blk.DarkColor` si el botón no tiene `textColor` propio
- Puede ser `<button>` o `<a>` (si tiene link)
- Inline style solo para: width, padding, border-radius

### Icon (`writeIconBlock`)

```html
<style>
  .{selector} .icon-wrap { color: {lightColor}; }
  :root[data-theme="dark"] .{selector} .icon-wrap { color: {darkColor}; }
</style>
<a class="b-link icon-wrap" href="...">
  <svg xmlns="..." width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor">
    <path d="..."/>
  </svg>
</a>
```

- 6092 iconos Tabler embebidos en `icons.json` via `go:embed` (1.5MB)
- `GetIconSVG(name, strokeWidth)` busca el icono y genera SVG inline
- Iconos **filled**: `fill="currentColor"` | Iconos **outline**: `stroke="currentColor"` con `stroke-width` configurable
- Tamaño fluido: `width="1em" height="1em"` escala con el font-size heredado
- Color light/dark con fallback a colores del site
- Si tiene enlace → `<a>` wrapper, si no → `<div>` wrapper

### Space (`writeSpaceBlock`)

```html
<div class="space-divider" style="border-top: {thick}px {mode} {color}"></div>
```

Divisor horizontal opcional dentro de un bloque de espacio.

### DarkMode (`writeDarkModeBlock`)

```html
<style>
  .{selector} .dm-toggle { color: {lightColor}; }
  :root[data-theme="dark"] .{selector} .dm-toggle { color: {darkColor}; }
</style>
<button class="dm-toggle" onclick="waxpToggleTheme()">
  <svg class="dm-icon-moon" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor">
    {IconMoonFilled path}
  </svg>
  <svg class="dm-icon-sun" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor">
    {IconSunFilled path}
  </svg>
</button>
```

- SVGs oficiales de Tabler Icons (`IconMoonFilled` / `IconSunFilled`) desde `icons.json`
- `font:inherit` hereda font-size fluido — SVGs con `1em` escalan proporcionalmente
- Color light/dark en `<style>` scoped (nunca inline)
- `user-select:none` para evitar selección
- El CSS base muestra luna en light y sol en dark

### LanguageSwitcher (`writeLangBlock`)

```html
<select class="lang-select" onchange="location.href=this.value">
  <option value="/nosotros" selected>ES</option>
  <option value="/en/about">EN</option>
</select>
```

- `font:inherit` hereda font-size fluido
- `appearance:none` sin caret nativo en desktop/tablet
- En mobile se resetea a `appearance:auto` con `font-size:16px!important` para que el dropdown nativo funcione
- Genera una opción por locale con URL calculada via `buildRoutePath`

### Menu (`writeMenuBlock`)

```html
<style>
  .{selector} { --m-color: #212529; --m-hover: #0d6efd; --m-active: #0a58ca; }
  :root[data-theme="dark"] .{selector} { --m-color: #f8f9fa; ... }
</style>
<nav class="menu-nav">
  <ul class="menu-list">
    <li class="menu-item">
      <a class="menu-link" href="/" style="font-family:...;font-weight:...">
        Inicio
      </a>
      <ul class="menu-sub">
        <li><a class="menu-sublink" href="/equipo">Equipo</a></li>
      </ul>
    </li>
  </ul>
</nav>
```

- Colores configurables: color, hover, active (light y dark)
- Font custom para nivel 1 (`menuFont`) y submenús (`menuSubFont`)
- Submenús desplegables via CSS `:hover` (`.menu-item:hover > .menu-sub { display: block }`)
- Soporta enlaces internos y externos
- Los items del menú vienen ya resueltos por locale desde `i18n.Resolve`

---

## Iconos embebidos (`icons.go`)

**Archivo:** `internal/render/icons.json` + `icons.go`

El fichero `icons.json` contiene 6092 iconos Tabler Icons mapeados por nombre:

```json
{
  "IconHomeFilled": { "type": "filled", "paths": ["M12.707 2.293..."] },
  "IconDog": { "type": "outline", "paths": ["M11 5h2", "M19 12..."] }
}
```

Se embebe en el binario via `go:embed`. `GetIconSVG(name, strokeWidth)` busca el icono y genera el SVG inline. El bloque `DarkMode` también usa esta fuente para la luna y el sol.

---

## Interacción con i18n

El renderizado consume datos **ya resueltos** por el paquete `i18n`:

| Campo en BD | Antes de Resolve | Después de Resolve |
|-------------|------------------|---------------------|
| `block.locales.text` | `{"es": "Hola", "en": "Hello"}` | `"Hola"` |
| `block.menu` | `{"es": [...], "en": [...]}` | `[...]` (array del locale) |
| `options.header.blocks[0].locales.label` | `{"es": "Inicio", "en": "Home"}` | `"Inicio"` |

`getLocale(locales, key)` en `html.go` extrae el string resuelto del mapa `locales`. En el caso del menú, ya llega como array de `MenuItem` con `label` como string directo.

---

## Estrategia de caché y regeneración

```
┌─────────────────────────────────────────────┐
│  PUT /api/sites/:id/live                    │
│  → ClearLiveSites + ActivateSiteLive        │
│  → renderPageForLocale × (pages × locales)  │
│  → UpsertPageRender                         │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  POST /api/sites/:id/regenerate             │
│  → renderPageForLocale × (pages × locales)  │
│  → UpsertPageRender                         │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  Petición pública (GET)                     │
│  → ServePage                                │
│  → GetPageRenderByPageAndLocale (lectura)   │
│  → devuelve HTML cacheado                   │
└─────────────────────────────────────────────┘
```

**Regeneración automática:** Al marcar un site como live (`PUT /api/sites/:id/live`) se regeneran todas sus páginas.

**Regeneración manual:** Via `POST /api/sites/:id/regenerate`.

**Regeneración pendiente:** Al actualizar una página o el site, habría que invalidar y regenerar los renders afectados. Esto incluiría:
- Al guardar una página → regenerar esa página para todos sus locales
- Al guardar el site (options/header/footer) → regenerar TODAS las páginas del site
- Al añadir/eliminar un locale → regenerar todas las páginas para el nuevo locale (o eliminar renders del locale borrado)
