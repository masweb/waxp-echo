# Locales API

Gestión de idiomas de un site. Todas las rutas requieren autenticación JWT.

```
Authorization: Bearer <token>
```

## Modelo

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | int64 | ID del locale |
| `site_id` | int64 | ID del site al que pertenece |
| `code` | string | Código de idioma (ISO 639-1), ej: `"es"`, `"en"`, `"ca"` |
| `is_default` | boolean | Indica si es el idioma por defecto del site |

**Reglas:**
- Un site puede tener **como máximo un** locale con `is_default: true`.
- No pueden repetirse `code` dentro del mismo site (`UNIQUE(site_id, code)`).
- Al añadir un locale con `is_default: true`, el locale que fuera default anterior pasa a `false` automáticamente.
- Al eliminar un locale, se eliminan en cascada los `blog_slugs`, `page_slugs` y `blocks` asociados.

---

## Add Locale to Site

Añade un nuevo idioma a un site existente.

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

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `code` | string | Sí | Código de idioma (ISO 639-1), máx 10 caracteres |
| `is_default` | boolean | No | Por defecto `false` |

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
| 400 | `code` vacío, site ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado |
| 409 | Código de locale duplicado para este site |

---

## Remove Locale from Site

Elimina un idioma de un site.

```
DELETE /api/sites/:id/locales/:localeId
```

**Response 204:** *(sin body)*

> El delete es en cascada: se eliminan los `blog_slugs`, `page_slugs` y `blocks` asociados a este locale.

**Errors:**
| Status | When |
|--------|------|
| 400 | Site ID o locale ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Site no encontrado, locale no encontrado |
