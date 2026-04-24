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

**Archivo:** `internal/handler/public.go:29`

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

## Regeneración de páginas (`RegenerateAllPages`)

**Archivo:** `internal/handler/public.go:121`

Endpoint que fuerza la regeneración de todas las páginas de un site:

```
POST /api/sites/:id/regenerate
```

Recorre todas las páginas publicadas × todos los locales y llama a `renderPageForLocale` para cada combinación. Devuelve:

```json
{ "regenerated": 12 }
```

---

## Función `renderPageForLocale`

**Archivo:** `internal/handler/public.go:163`

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
- **`SiteOptions`** — configuración global: colores, fuentes, breakpoints, dark mode, header/footer

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
- Variables CSS de tema: `--waxp-text`, `--waxp-bg`, `--waxp-ff`, `--waxp-fs`, `--waxp-lh`, `--waxp-dw`
- Variables dark: `:root[data-theme="dark"] { --waxp-text: ...; --waxp-bg: ... }`
- Estilos base de bloques: `.b`, `.b-inner`, `.b-overlay`, `.b-center`, `.b-link`
- Estilos de texto (tiptap): párrafos, imágenes, enlaces, listas, blockquotes
- Estilos de componentes: botones, dark mode toggle, language switcher, menús

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
- `hideOn` por breakpoint → `display: none !important`

**Media queries:**

| Breakpoint | Default | Configurable en |
|------------|---------|-----------------|
| Tablet | `≤1024px` | `options.tabletBP` |
| Mobile | `≤767px` | `options.mobileBP` |

Cada media query redefine grid, coordenadas de bloques y font-size.

**Tipografía fluida:**

Desktop: `font-size: calc(base_px + slope * vw)` con clamp en `targetWidth`.
Tablet/Mobile: `font-size: (fontSize + zoom)vw` donde zoom viene de `options.tabletTextZoom` / `options.mobileTextZoom`.

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
  
  <div class="b-inner">
    <!-- Contenido específico por tipo -->
  </div>
</div>
```

**Backgrounds con overlay:**

Los backgrounds de imagen se renderizan como overlays absolutos (`position: absolute; inset: 0`) con `pointer-events: none`. Si el tema dark tiene un overlay diferente, se inyecta un `<style>` scoped:

```html
<div class="b-overlay" style="background-image: url('...')"></div>
<style>:root[data-theme="dark"] .s1-b5 .b-overlay { background-image: url('...') }</style>
```

---

## Tipos de bloque

### Text (`writeTextBlock`)

```html
<div class="b-tiptap">{locales.text}</div>
```

El contenido de `locales.text` es HTML raw generado por TipTap (puede contener párrafos, listas, headings, imágenes, etc.).

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
<div class="btn-block">
  <a class="btn-b" href="..." style="background-color:...;color:...;--btn-hover-bg:...">
    {locales.label}
  </a>
  <style>
    .{selector} .btn-b:hover { background-color: var(--btn-hover-bg)!important; }
    :root[data-theme="dark"] .{selector} .btn-b { background-color: ...; }
  </style>
</div>
```

- Colores light/dark para: bg, text, hover, active, focus
- Estados con CSS custom properties (`--btn-hover-bg`, etc.)
- Puede ser `<button>` o `<a>` (si tiene link)
- Border radius, padding, width configurables

### Space (`writeSpaceBlock`)

```html
<div class="space-divider" style="border-top: {thick}px {mode} {color}"></div>
```

Divisor horizontal opcional dentro de un bloque de espacio.

### DarkMode (`writeDarkModeBlock`)

```html
<button class="dm-toggle" onclick="waxpToggleTheme()">
  <svg class="dm-icon-moon">...</svg>
  <svg class="dm-icon-sun">...</svg>
</button>
```

El CSS base muestra la luna en light y el sol en dark. El script `waxpToggleTheme()` alterna el atributo `data-theme` y lo persiste en `localStorage`.

### LanguageSwitcher (`writeLangBlock`)

```html
<select class="lang-select" onchange="location.href=this.value">
  <option value="/nosotros" selected>ES</option>
  <option value="/en/about">EN</option>
</select>
```

Genera una opción por locale con la URL calculada via `buildRoutePath`. Selecciona automáticamente el locale actual.

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
│  Guardar página/site (PUT/POST)             │
│  ↓                                          │
│  (aquí se engancharía la regeneración auto) │
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

**Estado actual:** La regeneración es manual vía `RegenerateAllPages`. No existe todavía un trigger automático al guardar páginas o sitios.

**Regeneración pendiente:** Al actualizar una página o el site, habría que invalidar y regenerar los renders afectados. Esto incluiría:
- Al guardar una página → regenerar esa página para todos sus locales
- Al guardar el site (options/header/footer) → regenerar TODAS las páginas del site
- Al añadir/eliminar un locale → regenerar todas las páginas para el nuevo locale (o eliminar renders del locale borrado)
