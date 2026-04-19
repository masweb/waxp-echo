# Sites API

Todas las rutas requieren autenticación JWT.

```
Authorization: Bearer <token>
```

---

## Create Site

Crea un site con sus páginas por defecto (raíz `/` y `404`) en una sola transacción. Requiere al menos un locale.

```
POST /api/sites
```

**Body:**
```json
{
  "name": "Mi Blog",
  "domain": "miblog.com",
  "options": {},
  "locales": [
    { "code": "es", "is_default": true },
    { "code": "en", "is_default": false }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `name` | string | Sí | Nombre del site (máx 255 caracteres) |
| `domain` | string | Sí | Dominio del site (único) |
| `options` | object | No | Opciones del site (JSON). Por defecto `{}` |
| `locales` | array | **Sí** | Al menos un locale es requerido |

**Objeto locale:**

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `code` | string | Código de idioma (ISO 639-1), ej: `"es"`, `"en"`, `"ca"` |
| `is_default` | boolean | `true` para marcar como idioma por defecto del site |

**Reglas:**
- Un site debe tener **como máximo un** locale con `is_default: true`.
- No pueden repetirse `code` dentro del mismo site.
- Se crean automáticamente dos páginas publicadas:
  - **Raíz** (`slug: ""`) — una entrada por cada locale.
  - **404** (`slug: "404"`) — una entrada por cada locale, mismo nombre en todos los idiomas.
- Toda la operación (site + locales + páginas + slugs + SEO) es atómica.

**Response 201:**
```json
{
  "id": 1,
  "name": "Mi Blog",
  "domain": "miblog.com",
  "options": {},
  "locales": [
    { "code": "es", "is_default": true },
    { "code": "en", "is_default": false }
  ],
  "pages": [
    { "id": 1, "site_id": 1, "type": "page", "slug": "", "locale_code": "es" },
    { "id": 1, "site_id": 1, "type": "page", "slug": "", "locale_code": "en" },
    { "id": 2, "site_id": 1, "type": "page", "slug": "404", "locale_code": "es" },
    { "id": 2, "site_id": 1, "type": "page", "slug": "404", "locale_code": "en" }
  ]
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Body inválido, name o domain vacíos, locales vacío, locale sin code, más de un is_default, options JSON inválido |
| 401 | Token missing, invalid or expired |
| 409 | Domain ya existe, código de locale duplicado |

---

## List Sites

```
GET /api/sites
```

**Query Params:**
| Param | Type | Max | Description |
|-------|------|-----|-------------|
| `cursor` | int64 | - | ID del último elemento recibido. Omite para la primera página |
| `limit` | int32 | 100 | Cantidad de elementos. **Opcional**: sin él retorna todos los registros |

Ver [Paginación](./pagination.md) para detalle del comportamiento.

**Response 200:**
```json
{
  "data": [
      {
        "id": 1,
        "name": "Mi Blog",
        "domain": "miblog.com",
        "options": {},
        "locales": [
          { "code": "es", "is_default": true },
          { "code": "en", "is_default": false }
        ]
      },
      {
        "id": 2,
        "name": "Tienda",
        "domain": "tienda.com",
        "options": {},
        "locales": [
          { "code": "es", "is_default": true }
        ]
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
| 400 | `cursor` o `limit` inválidos |
| 401 | Token missing, invalid or expired |

---

## Get Site

```
GET /api/sites/:id
```

**Response 200:**
```json
{
  "id": 1,
  "name": "Mi Blog",
  "domain": "miblog.com",
  "options": {},
  "locales": [
    { "code": "es", "is_default": true },
    { "code": "en", "is_default": false }
  ],
  "routes": {
    "es": [
      { "path": "/sobre-nosotros", "page_id": 1 }
    ],
    "en": [
      { "path": "/en/about", "page_id": 1 }
    ]
  }
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

---

## Update Site

```
PUT /api/sites/:id
```

**Body:**
```json
{
  "name": "Mi Blog v2",
  "domain": "miblogv2.com",
  "options": { "theme": "dark" }
}
```

**Response 200:**
```json
{
  "id": 1,
  "name": "Mi Blog v2",
  "domain": "miblogv2.com",
  "options": { "theme": "dark" },
  "locales": [
    { "code": "es", "is_default": true },
    { "code": "en", "is_default": false }
  ]
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Body inválido, ID inválido, name o domain vacíos, options JSON inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |
| 409 | Domain ya existe (en otro site) |

---

## Add Locale to Site

```
POST /api/sites/:id/locales
```

**Body:**
```json
{
  "code": "ca",
  "is_default": false
}
```

**Response 201:**
```json
{
  "code": "ca",
  "is_default": false
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | `code` vacío, site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |
| 409 | Código de locale duplicado para este site |

---

## Remove Locale from Site

```
DELETE /api/sites/:id/locales/:localeCode
```

**Response 204:** *(sin body)*

> El delete es en cascada: se eliminan los `blog_slugs`, `page_slugs` y `blocks` asociados a este locale.

**Errors:**
| Status | When |
|--------|------|
| 400 | ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site o locale no encontrado |

---

## Delete Site

```
DELETE /api/sites/:id
```

**Response 204:** *(sin body)*

**Errors:**
| Status | When |
|--------|------|
| 400 | ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |

> El delete es en cascada: se eliminan los `site_locales`, `blogs`, `pages`, `blocks` y tablas relacionadas del site.
