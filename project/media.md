# Media API

Todas las rutas requieren autenticación JWT, excepto la ruta pública de servir ficheros.

```
Authorization: Bearer <token>
```

---

## Upload Media

Sube un archivo de imagen al servidor. El archivo se almacena en disco (`MEDIA_DIR`) y se registra en la base de datos.

```
POST /api/media
```

**Content-Type:** `multipart/form-data`

**Body (form-data):**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `file` | file | Sí | Archivo de imagen |

**MIME types permitidos:**
- `image/jpeg`
- `image/png`
- `image/gif`
- `image/webp`
- `image/svg+xml`

**Response 201:**
```json
{
  "id": 1,
  "filename": "foto.jpg",
  "mime_type": "image/jpeg",
  "size": 204800,
  "url": "/media/1713500000000000000.jpg",
  "created_at": "2026-04-19T12:00:00Z"
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | No se envió archivo, MIME type no permitido |
| 401 | Token missing, invalid or expired |

---

## List Media

Lista los archivos multimedia paginados, ordenados por fecha de creación descendente.

```
GET /api/media
```

**Query Params:**
| Param | Type | Default | Max | Description |
|-------|------|---------|-----|-------------|
| `page` | int | 1 | - | Número de página |
| `per_page` | int | 20 | 100 | Elementos por página |

**Response 200:**
```json
{
  "data": [
    {
      "id": 2,
      "filename": "logo.png",
      "mime_type": "image/png",
      "size": 51200,
      "url": "/media/1713500000000000001.png",
      "created_at": "2026-04-19T12:01:00Z"
    },
    {
      "id": 1,
      "filename": "foto.jpg",
      "mime_type": "image/jpeg",
      "size": 204800,
      "url": "/media/1713500000000000000.jpg",
      "created_at": "2026-04-19T12:00:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "per_page": 20
}
```

**Errors:**
| Status | When |
|--------|------|
| 401 | Token missing, invalid or expired |

---

## Get Media

Obtiene los metadatos de un archivo multimedia por su ID.

```
GET /api/media/:id
```

**Response 200:**
```json
{
  "id": 1,
  "filename": "foto.jpg",
  "mime_type": "image/jpeg",
  "size": 204800,
  "url": "/media/1713500000000000000.jpg",
  "created_at": "2026-04-19T12:00:00Z"
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Media no encontrado |

---

## Delete Media

Elimina un archivo multimedia. Borra tanto el registro de la base de datos como el fichero en disco.

```
DELETE /api/media/:id
```

**Response 204:** *(sin body)*

**Errors:**
| Status | When |
|--------|------|
| 400 | ID inválido |
| 401 | Token missing, invalid or expired |
| 404 | Media no encontrado |

---

## Serve Media (pública)

Sirve un archivo multimedia desde disco. No requiere autenticación.

```
GET /media/:name
```

**Response 200:** El archivo binario con el `Content-Type` correspondiente.

**Errors:**
| Status | When |
|--------|------|
| 400 | Nombre vacío |
| 404 | Archivo no encontrado en disco |
