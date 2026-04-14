# Sites API

Todas las rutas requieren autenticación JWT.

```
Authorization: Bearer <token>
```

---

## Create Site

```
POST /api/sites
```

**Body:**
```json
{
  "name": "Mi Blog",
  "domain": "miblog.com",
  "locales": [
    { "code": "es", "is_default": true },
    { "code": "en", "is_default": false }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `name` | string | Sí | Nombre del site |
| `domain` | string | Sí | Dominio del site (único) |
| `locales` | array | No | Idiomas del site. Si se omite, se crea sin locales |

**Objeto locale:**

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `code` | string | Código de idioma (ISO 639-1), ej: `"es"`, `"en"`, `"ca"` |
| `is_default` | boolean | `true` para marcar como idioma por defecto del site |

**Reglas:**
- Un site debe tener **como máximo un** locale con `is_default: true`.
- No pueden repetirse `code` dentro del mismo site.
- Los locales se insertan en la tabla `site_locales` junto con la creación del site, en una **transacción atómica**. Si falla la inserción de locales, no se crea el site.

**Response 201:**
```json
{
  "id": 1,
  "name": "Mi Blog",
  "domain": "miblog.com",
  "locales": [
    { "id": 1, "code": "es", "is_default": true },
    { "id": 2, "code": "en", "is_default": false }
  ]
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Body inválido, name o domain vacíos, locale sin code, más de un is_default |
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
      "locales": [
        { "id": 1, "code": "es", "is_default": true },
        { "id": 2, "code": "en", "is_default": false }
      ]
    },
    {
      "id": 2,
      "name": "Tienda",
      "domain": "tienda.com",
      "locales": [
        { "id": 3, "code": "es", "is_default": true }
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
  "locales": [
    { "id": 1, "code": "es", "is_default": true },
    { "id": 2, "code": "en", "is_default": false }
  ]
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
  "domain": "miblogv2.com"
}
```

**Response 200:**
```json
{
  "id": 1,
  "name": "Mi Blog v2",
  "domain": "miblogv2.com",
  "locales": [
    { "id": 1, "code": "es", "is_default": true },
    { "id": 2, "code": "en", "is_default": false }
  ]
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Body inválido, ID inválido, name o domain vacíos |
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
  "id": 3,
  "code": "ca",
  "is_default": false
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | `code` vacío, más de un `is_default: true` en el site |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |
| 409 | Código de locale duplicado para este site |

---

## Remove Locale from Site

```
DELETE /api/sites/:id/locales/:localeId
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
