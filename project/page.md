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
| `blog_id` | int64\|null | ID del blog (solo para posts) |
| `parent_id` | int64\|null | ID de la página padre (anidación) |
| `type` | string | `"page"` o `"post"` |
| `layout` | JSONB | Contenido/layout de la página |
| `published_at` | string\|null | Fecha de publicación (ISO 8601). `null` = borrador |
| `slugs` | array | Slugs por locale |
| `created_at` | string | Fecha de creación (ISO 8601) |
| `updated_at` | string | Fecha de actualización (ISO 8601) |

### Slug

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | int64 | ID del slug |
| `locale_id` | int64 | ID del locale |
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
- Al eliminar una página se eliminan en cascada sus slugs y páginas hijas.

---

## Create Page

```
POST /api/sites/:id/pages
```

**Body (page):**
```json
{
  "type": "page",
  "parent_id": null,
  "layout": {},
  "published_at": "2026-04-14T10:00:00Z",
  "slugs": [
    { "locale_id": 1, "slug": "about" },
    { "locale_id": 2, "slug": "sobre-nosotros" }
  ]
}
```

**Body (post):**
```json
{
  "type": "post",
  "blog_id": 1,
  "layout": {},
  "published_at": "2026-04-14T10:00:00Z",
  "slugs": [
    { "locale_id": 1, "slug": "mi-articulo" },
    { "locale_id": 2, "slug": "my-article" }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `type` | string | Sí | `"page"` o `"post"` |
| `blog_id` | int64\|null | Sí si `type="post"` | ID del blog. Debe pertenecer al site |
| `parent_id` | int64\|null | No | ID de la página padre. Mismo tipo y (para posts) mismo blog |
| `layout` | object | No | Layout de la página. Por defecto `{}` |
| `published_at` | string\|null | No | Fecha de publicación (ISO 8601). `null` o omitir = borrador |
| `slugs` | array | Sí | Al menos un slug. Cada `locale_id` debe pertenecer al site |

**Response 201:**
```json
{
  "id": 1,
  "site_id": 1,
  "blog_id": null,
  "parent_id": null,
  "type": "page",
  "layout": {},
  "published_at": "2026-04-14T10:00:00Z",
  "slugs": [
    { "id": 1, "locale_id": 1, "slug": "about" },
    { "id": 2, "locale_id": 2, "slug": "sobre-nosotros" }
  ],
  "created_at": "2026-04-14T10:00:00Z",
  "updated_at": "2026-04-14T10:00:00Z"
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | `type` inválido, sin slugs, slug vacío, locale_id no pertenece al site, `blog_id` requerido para posts, blog no encontrado, parent_id no encontrado o tipo incompatible, página no puede ser su propio padre |
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
      "slugs": [
        { "id": 1, "locale_id": 1, "slug": "about" },
        { "id": 2, "locale_id": 2, "slug": "sobre-nosotros" }
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

**Errors:**
| Status | When |
|--------|------|
| 400 | `cursor` o `limit` inválidos, site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

---

## Get Page

```
GET /api/sites/:id/pages/:pageId
```

**Response 200:** Mismo formato que Create Page response.

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID o page ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

---

## Update Page

```
PUT /api/sites/:id/pages/:pageId
```

El `type` y `blog_id` no se pueden modificar tras la creación. Solo se actualizan `parent_id`, `layout`, `published_at` y `slugs`.

**Body:**
```json
{
  "parent_id": 1,
  "layout": { "components": [...] },
  "published_at": "2026-04-14T12:00:00Z",
  "slugs": [
    { "locale_id": 1, "slug": "about-us" },
    { "locale_id": 2, "slug": "sobre-nosotros" }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `parent_id` | int64\|null | No | Cambiar padre. Mismo tipo y (para posts) mismo blog. La página no puede ser su propio padre |
| `layout` | object | No | Actualizar layout. Por defecto `{}` |
| `published_at` | string\|null | No | Actualizar fecha de publicación |
| `slugs` | array | Sí | Reemplaza todos los slugs existentes. Al menos uno requerido |

> Los slugs se eliminan y reinsertan en una transacción atómica.

**Response 200:** Mismo formato que Get Page.

**Errors:**
| Status | When |
|--------|------|
| 400 | Sin slugs, slug vacío, locale_id no pertenece al site, parent_id inválido o tipo incompatible, página no puede ser su propio padre |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

---

## Delete Page

```
DELETE /api/sites/:id/pages/:pageId
```

**Response 204:** *(sin body)*

> El delete es en cascada: se eliminan los `page_slugs` asociados y todas las páginas hijas (y sus slugs, recursivamente).

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID o page ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Página no encontrada |

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
