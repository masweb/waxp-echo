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
  "domain": "miblog.com"
}
```

**Response 201:**
```json
{
  "id": 1,
  "name": "Mi Blog",
  "domain": "miblog.com"
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Body inválido, name o domain vacíos |
| 401 | Token missing, invalid or expired |
| 409 | Domain ya existe |

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
      "domain": "miblog.com"
    },
    {
      "id": 2,
      "name": "Tienda",
      "domain": "tienda.com"
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
  "domain": "miblog.com"
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
  "domain": "miblogv2.com"
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
