# Pages API

Gestión de páginas y posts de un site. Todas las rutas requieren autenticación JWT.

```
Authorization: Bearer <token>
```

## Modelo

### Page

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | int64 | ID de la página |
| `site_id` | int64 | ID del site al que pertenece |
| `blog_id` | int64|null | ID del blog (solo para posts) |
| `parent_id` | int64|null | ID de la página padre (anidación) |
| `type` | string | `"page"` o `"post"` |
| `layout` | JSONB | Contenido/layout de la página |
| `published_at` | string|null | Fecha de publicación (ISO 8601). `null` = borrador |
| `seo` | array | SEO por locale |
| `slugs` | array | Slugs por locale |
| `created_at` | string | Fecha de creación (ISO 8601) |
| `updated_at` | string | Fecha de actualización (ISO 8601) |

### PageSeo

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | int64 | ID del registro SEO |
| `locale_code` | string | Código de idioma (ISO 639-1) |
| `title` | string | Título de la página (SEO) |
| `description` | string | Descripción (SEO meta description) |

### Slug

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | int64 | ID del slug |
| `locale_code` | string | Código de idioma (ISO 639-1) |
| `slug` | string | Segmento de URL (no la ruta completa) |

### Tipos de página

| Tipo | Descripción | `blog_id` | `parent_id` |
|------|-------------|-----------|-------------|
| `page` | Página normal del CMS | `null` | Otra página tipo `page` |
| `post` | Entrada de blog | Requerido (FK a `blogs`) | Otro post del mismo blog, o `null` |

Los posts pueden anidarse vía `parent_id` para crear sub-posts (ej: artículos multi-parte).

**Estructura:**
```
blogs (tabla separada, con blog_slugs propios)
├── Post "Artículo 1" (type='post', blog_id=X, parent_id=NULL)
│   ├── Post "Parte 1" (type='post', blog_id=X, parent_id=Art1)
│   └── Post "Parte 2" (type='post', blog_id=X, parent_id=Art1)
└── Post "Artículo 2" (type='post', blog_id=X, parent_id=NULL)

Pages (type='page', blog_id=NULL)
├── Page "Inicio" (parent_id=NULL)
├── Page "Sobre nosotros" (parent_id=NULL)
│   └── Page "Equipo" (parent_id=Sobre)
└── Page "Contacto" (parent_id=NULL)
```

**Reglas generales:**
- Las páginas pueden anidarse infinitamente vía `parent_id`.
- `parent_id` debe ser del mismo tipo: page→page, post→post.
- Para posts, el padre debe pertenecer al mismo blog.
- Cada slug es un segmento individual. La ruta completa se compone concatenando los slugs de la cadena de ancestros.
- Solo las páginas con `published_at` no nulo aparecen en el endpoint de rutas.
- Al eliminar una página se eliminan en cascada sus slugs, SEO y páginas hijas.

---

## Layout y locales

El campo `layout` es un JSONB que contiene secciones y bloques. Los bloques pueden tener un campo `locales` con campos traducibles por idioma.

### Almacenamiento en BD

Cada valor dentro de `locales` es un mapa `{locale_code: string}`:

```json
[
  {
    "id": 1,
    "mobile": { "cols": 8, "rows": 12 },
    "tablet": { "cols": 12, "rows": 16 },
    "desktop": { "cols": 24, "rows": 20 },
    "blocks": [
      {
        "id": 10,
        "type": "heading",
        "locales": {
          "title": { "es": "Sobre nosotros", "en": "About us" },
          "subtitle": { "es": "Conoce nuestro equipo", "en": "Meet our team" }
        },
        "config": { "level": 2 }
      },
      {
        "id": 11,
        "type": "paragraph",
        "locales": {
          "text": { "es": "Bienvenidos", "en": "Welcome" }
        }
      },
      {
        "id": 12,
        "type": "image",
        "config": { "src": "/img.jpg", "alt": "Foto" }
      }
    ]
  }
]
```

### Comportamiento por idioma

Las operaciones de lectura y escritura requieren el query param `?locale=es`. El backend se encarga de resolver las traducciones de forma transparente:

- **Lectura (GET)**: Los valores de `locales` se resuelven a strings del idioma solicitado. Si un locale no tiene traducción para un campo, se devuelve `""`.
- **Creación (POST)**: Los strings recibidos se envuelven en el mapa de idiomas antes de guardar.
- **Actualización (PUT)**: Los strings entrantes se mergean en el mapa existente. Los demás idiomas se preservan intactos. El merge se hace por `id` de bloque/sección.

> Los bloques sin campo `locales` (como `image` en el ejemplo) no se ven afectados.

---

## Create Page

```
POST /api/sites/:id/pages?locale=es
```

**Query Params:**
| Param | Type | Requerido | Descripción |
|-------|------|-----------|-------------|
| `locale` | string | Sí | Código de idioma (ISO 639-1). Debe pertenecer al site |

**Body (page con layout):**
```json
{
  "type": "page",
  "parent_id": null,
  "layout": [
    {
      "id": 1,
      "mobile": { "cols": 8, "rows": 12 },
      "tablet": { "cols": 12, "rows": 16 },
      "desktop": { "cols": 24, "rows": 20 },
      "blocks": [
        {
          "id": 10,
          "type": "heading",
          "locales": {
            "title": "Sobre nosotros",
            "subtitle": "Conoce nuestro equipo"
          },
          "config": { "level": 2 }
        }
      ]
    }
  ],
  "published_at": "2026-04-14T10:00:00Z",
  "seo": [
    { "locale_code": "es", "title": "Sobre nosotros", "description": "Conoce nuestro equipo" }
  ],
  "slugs": [
    { "locale_code": "es", "slug": "sobre-nosotros" },
    { "locale_code": "en", "slug": "about" }
  ]
}
```

> Los valores dentro de `locales` son strings simples (no mapas). El backend los envuelve en `{ "es": "..." }` al guardar en BD.

**Body (post):**
```json
{
  "type": "post",
  "blog_id": 1,
  "layout": {},
  "published_at": "2026-04-14T10:00:00Z",
  "seo": [
    { "locale_code": "es", "title": "Mi artículo", "description": "Descripción SEO" }
  ],
  "slugs": [
    { "locale_code": "es", "slug": "mi-articulo" },
    { "locale_code": "en", "slug": "my-article" }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `type` | string | Sí | `"page"` o `"post"` |
| `blog_id` | int64|null | Sí si `type="post"` | ID del blog. Debe pertenecer al site |
| `parent_id` | int64|null | No | ID de la página padre. Mismo tipo y (para posts) mismo blog |
| `layout` | object | No | Layout de la página. Si está vacío o ausente, se generan 4 secciones iniciales con IDs únicos |
| `published_at` | string|null | No | Fecha de publicación (ISO 8601). `null` o omitir = borrador |
| `seo` | array | No | SEO titles y descriptions por locale |
| `slugs` | array | Sí | Al menos un slug. Cada `locale_code` debe pertenecer al site |

**Layout por defecto:**
Cuando `layout` está vacío, `{}`, o `null`, el backend genera automáticamente 4 secciones:

```json
[
  {
    "id": 1,
    "mobile": { "cols": 8, "rows": 12 },
    "tablet": { "cols": 12, "rows": 16 },
    "desktop": { "cols": 24, "rows": 20 },
    "blocks": []
  },
  { "id": 2, "mobile": {...}, "tablet": {...}, "desktop": {...}, "blocks": [] },
  { "id": 3, "mobile": {...}, "tablet": {...}, "desktop": {...}, "blocks": [] },
  { "id": 4, "mobile": {...}, "tablet": {...}, "desktop": {...}, "blocks": [] }
]
```

> Los IDs de sección son únicos globalmente por site y se asignan desde un contador persistente.

**Response 201:**
```json
{
  "id": 1,
  "site_id": 1,
  "blog_id": null,
  "parent_id": null,
  "type": "page",
  "layout": [
    {
      "id": 1,
      "mobile": { "cols": 8, "rows": 12 },
      "tablet": { "cols": 12, "rows": 16 },
      "desktop": { "cols": 24, "rows": 20 },
      "blocks": [
        {
          "id": 10,
          "type": "heading",
          "locales": {
            "title": "Sobre nosotros",
            "subtitle": "Conoce nuestro equipo"
          },
          "config": { "level": 2 }
        }
      ]
    }
  ],
  "published_at": "2026-04-14T10:00:00Z",
  "seo": [
    { "id": 1, "locale_code": "es", "title": "Sobre nosotros", "description": "Conoce nuestro equipo" }
  ],
  "slugs": [
    { "id": 1, "locale_code": "es", "slug": "sobre-nosotros" },
    { "id": 2, "locale_code": "en", "slug": "about" }
  ],
  "created_at": "2026-04-14T10:00:00Z",
  "updated_at": "2026-04-14T10:00:00Z"
}
```

> La respuesta devuelve el layout resuelto para el locale solicitado (strings, no mapas).

**Errors:**
| Status | When |
|--------|------|
| 400 | `type` inválido, sin slugs, slug vacío, locale_code no pertenece al site, `blog_id` requerido para posts, blog no encontrado, parent_id no encontrado o tipo incompatible, página no puede ser su propio padre, seo sin locale_code o locale_code no pertenece al site, locale requerido, locale no pertenece al site |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

---

## List Pages

```
GET /api/sites/:id/pages
```

**Query Params:**
| Param | Type | Max | Description |
|-------|------|-----|-------------|
| `cursor` | int64 | - | ID del último elemento recibido. Omite para la primera página |
| `limit` | int32 | 100 | Cantidad de elementos. **Opcional**: sin él retorna todos los registros |

**Filtros disponibles:**
| Param | Descripción |
|-------|-------------|
| `filter[type]` | Filtrar por tipo: `"page"` o `"post"` |
| `filter[parent_id]` | Filtrar por parent_id exacto |
| `filter[blog_id]` | Filtrar por blog_id |
| `filter[published_at_isnull]` | `true` = borradores, `false` = publicadas |
| `filter[id]` | Filtrar por ID |

Ver [Paginación](./pagination.md) para detalle del comportamiento.

**Response 200:**
```json
{
  "data": [
    {
      "id": 1,
      "site_id": 1,
      "blog_id": null,
      "parent_id": null,
      "type": "page",
      "layout": {},
      "published_at": "2026-04-14T10:00:00Z",
      "seo": [
        { "id": 1, "locale_code": "es", "title": "Sobre nosotros", "description": "Contenido SEO" },
        { "id": 2, "locale_code": "en", "title": "About us", "description": "SEO content" }
      ],
      "slugs": [
        { "id": 1, "locale_code": "es", "slug": "sobre-nosotros" },
        { "id": 2, "locale_code": "en", "slug": "about" }
      ],
      "created_at": "2026-04-14T10:00:00Z",
      "updated_at": "2026-04-14T10:00:00Z"
    }
  ],
  "next_cursor": 1,
  "total": 5,
  "has_more": true
}
```

> El listado devuelve el layout sin resolver (formato BD). No requiere `?locale`.

**Errors:**
| Status | When |
|--------|------|
| 400 | `cursor` o `limit` inválidos, site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

---

## Get Page

```
GET /api/sites/:id/pages/:pageId?locale=es
```

**Query Params:**
| Param | Type | Requerido | Descripción |
|-------|------|-----------|-------------|
| `locale` | string | Sí | Código de idioma (ISO 639-1). Debe pertenecer al site |

**Response 200:**
```json
{
  "id": 1,
  "site_id": 1,
  "blog_id": null,
  "parent_id": null,
  "type": "page",
  "layout": [
    {
      "id": 1,
      "mobile": { "cols": 8, "rows": 12 },
      "tablet": { "cols": 12, "rows": 16 },
      "desktop": { "cols": 24, "rows": 20 },
      "blocks": [
        {
          "id": 10,
          "type": "heading",
          "locales": {
            "title": "Sobre nosotros",
            "subtitle": "Conoce nuestro equipo"
          },
          "config": { "level": 2 }
        }
      ]
    }
  ],
  "published_at": "2026-04-14T10:00:00Z",
  "seo": [
    { "id": 1, "locale_code": "es", "title": "Sobre nosotros", "description": "Contenido SEO" },
    { "id": 2, "locale_code": "en", "title": "About us", "description": "SEO content" }
  ],
  "slugs": [
    { "id": 1, "locale_code": "es", "slug": "sobre-nosotros" },
    { "id": 2, "locale_code": "en", "slug": "about" }
  ],
  "created_at": "2026-04-14T10:00:00Z",
  "updated_at": "2026-04-14T10:00:00Z"
}
```

> Si el locale es `en` y un bloque solo tiene `es`, los campos de `locales` se devuelven como `""`.

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID o page ID inválido, locale requerido, locale no pertenece al site |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

---

## Update Page

```
PUT /api/sites/:id/pages/:pageId?locale=es
```

El `type` y `blog_id` no se pueden modificar tras la creación. Solo se actualizan `parent_id`, `layout`, `published_at`, `seo` y `slugs`.

**Query Params:**
| Param | Type | Requerido | Descripción |
|-------|------|-----------|-------------|
| `locale` | string | Sí | Código de idioma (ISO 639-1). Debe pertenecer al site |

**Body:**
```json
{
  "parent_id": 1,
  "layout": [
    {
      "id": 1,
      "mobile": { "cols": 8, "rows": 12 },
      "tablet": { "cols": 12, "rows": 16 },
      "desktop": { "cols": 24, "rows": 20 },
      "blocks": [
        {
          "id": 10,
          "type": "heading",
          "locales": {
            "title": "Sobre nosotros - editado",
            "subtitle": "Conoce nuestro equipo"
          },
          "config": { "level": 2 }
        }
      ]
    }
  ],
  "published_at": "2026-04-14T12:00:00Z",
  "seo": [
    { "locale_code": "es", "title": "Sobre nosotros", "description": "Conoce nuestro equipo" },
    { "locale_code": "en", "title": "About us", "description": "Meet our team" }
  ],
  "slugs": [
    { "locale_code": "es", "slug": "about-us" },
    { "locale_code": "en", "slug": "sobre-nosotros" }
  ]
}
```

> Los valores dentro de `locales` son strings simples. El backend los mergea en el mapa existente por `id` de bloque, preservando los demás idiomas.

**Ejemplo de merge en BD:**

Si el bloque 10 tenía en BD:
```json
{ "title": { "es": "Sobre nosotros", "en": "About us" } }
```
Y se envía `PUT ?locale=es` con:
```json
{ "title": "Sobre nosotros - editado" }
```
El resultado en BD será:
```json
{ "title": { "es": "Sobre nosotros - editado", "en": "About us" } }
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `parent_id` | int64|null | No | Cambiar padre. Mismo tipo y (para posts) mismo blog. La página no puede ser su propio padre |
| `layout` | object | No | Actualizar layout. Los `locales` se mergean por idioma, el resto se reemplaza |
| `published_at` | string|null | No | Actualizar fecha de publicación |
| `seo` | array | No | Reemplaza todos los SEO. Cada `locale_code` debe pertenecer al site |
| `slugs` | array | Sí | Reemplaza todos los slugs existentes. Al menos uno requerido |

> Los slugs y SEO se eliminan y reinsertan en una transacción atómica.

**Response 200:** Mismo formato que Get Page (layout resuelto para el locale solicitado).

**Errors:**
| Status | When |
|--------|------|
| 400 | Sin slugs, slug vacío, locale_code no pertenece al site, parent_id inválido o tipo incompatible, página no puede ser su propio padre, seo sin locale_code o locale_code no pertenece al site, locale requerido, locale no pertenece al site |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

---

## Delete Page

```
DELETE /api/sites/:id/pages/:pageId
```

**Response 204:** *(sin body)*

> El delete es en cascada: se eliminan los `page_slugs` y `page_seo` asociados y todas las páginas hijas (y sus slugs y SEO, recursivamente).

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID o page ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

---

## List Page Revisions

Lista las revisiones de una página de forma paginada. Cada vez que se crea o actualiza el `layout` de una página se genera automáticamente una revisión.

```
GET /api/sites/:id/pages/:pageId/revisions
```

**Query params:**

| Param | Type | Description |
|-------|------|-------------|
| `cursor` | int64 | ID de la última revisión de la página anterior |
| `limit` | int32 | Cantidad de elementos (max 100). Sin limite = todos |
| `filter[id]` | int64 | Filtrar por ID |
| `filter[revision_number]` | int32 | Filtrar por número de revisión |

**Response 200:**
```json
{
  "data": [
    {
      "id": 1,
      "revision_number": 1,
      "created_at": "2026-04-22T10:00:00Z"
    },
    {
      "id": 2,
      "revision_number": 2,
      "created_at": "2026-04-22T11:30:00Z"
    }
  ],
  "next_cursor": 2,
  "total": 15,
  "has_more": true
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID o page ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

---

## Restore Page Revision

Restaura el `layout` de una revisión anterior en la página actual. El layout actual se guarda automáticamente como revisión antes de sobreescribirse (trigger de BD), por lo que siempre se puede volver atrás restaurando la revisión anterior.

```
POST /api/sites/:id/pages/:pageId/revisions/:revisionId/restore?locale=es
```

**Response 200:** *(devuelve la página completa con el layout restaurado, procesado para el locale solicitado)*
```json
{
  "id": 5,
  "site_id": 1,
  "blog_id": null,
  "parent_id": null,
  "type": "page",
  "layout": { "...": "layout restaurado con solo el locale solicitado" },
  "published_at": "2026-04-22T10:00:00Z",
  "seo": [],
  "slugs": [],
  "created_at": "2026-04-22T10:00:00Z",
  "updated_at": "2026-04-22T12:00:00Z"
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID, page ID o revision ID inválido, locale requerido, layout idéntico al actual |
| 401 | Token missing, invalid or expired |
| 404 | Página o revisión no encontrada |

---

## Routes

Devuelve el mapa completo de rutas por idioma para un site. Incluye páginas, blogs y posts publicados. Diseñado para cargar en vue-router.

```
GET /api/sites/:id/routes
```

**Reglas de rutas:**
- **Pages**: La ruta se compone recorriendo la cadena de ancestros por sus slugs.
- **Blogs**: Usan sus `blog_slugs` como ruta base.
- **Posts**: La ruta es `{blog_slug}/{post_slug_chain}`, concatenando el slug del blog con los slugs del post y sus ancestros.
- El locale por defecto **no lleva prefijo** de idioma: `/about/team`
- Los locales no por defecto llevan prefijo `/{code}`: `/en/about/team`
- Solo se incluyen páginas y posts con `published_at` no nulo.

**Response 200:**
```json
{
  "routes": {
    "es": [
      { "path": "/sobre-nosotros", "page_id": 1 },
      { "path": "/sobre-nosotros/equipo", "page_id": 2 },
      { "path": "/contacto", "page_id": 3 },
      { "path": "/noticias", "blog_id": 1 },
      { "path": "/noticias/mi-articulo", "page_id": 4 },
      { "path": "/noticias/mi-articulo/parte-1", "page_id": 5 }
    ],
    "en": [
      { "path": "/en/about", "page_id": 1 },
      { "path": "/en/about/team", "page_id": 2 },
      { "path": "/en/contact", "page_id": 3 },
      { "path": "/en/news", "blog_id": 1 },
      { "path": "/en/news/my-article", "page_id": 4 },
      { "path": "/en/news/my-article/part-1", "page_id": 5 }
    ]
  }
}
```

**RouteEntry:**
| Campo | Tipo | Descripción |
|-------|------|-------------|
| `path` | string | Ruta completa |
| `page_id` | int64 | Presente si es una página o post |
| `blog_id` | int64 | Presente si es un blog (listado) |

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

---

## Next Section ID

Obtiene el siguiente ID de sección único para un site. Útil cuando el frontend necesita crear nuevas secciones y requiere un ID único del backend.

```
POST /api/sites/:id/sections/next-id
```

**Response 200:**
```json
{
  "id": 5
}
```

> El contador de IDs es global por site y se persiste en la base de datos. Cada llamada incrementa el contador.

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

---

## Next Block ID

Obtiene el siguiente ID de bloque único para un site. Útil cuando el frontend necesita crear nuevos bloques y requiere un ID único del backend.

```
POST /api/sites/:id/blocks/next-id
```

**Response 200:**
```json
{
  "id": 5
}
```

> El contador de IDs es global por site y se persiste en la base de datos. Cada llamada incrementa el contador. Funciona igual que el contador de secciones: usa un `upsert` atómico sobre la tabla `block_counters`.

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |
