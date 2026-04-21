# Media API

Todas las rutas requieren autenticación JWT, excepto la ruta pública de servir ficheros.

```
Authorization: Bearer <token>
```

---

## Thumbnails

Al subir una imagen se genera automáticamente un thumbnail de 150x150px con recorte tipo **cover** (center crop). El thumbnail se guarda en el mismo directorio que la imagen original.

**Reglas:**
- Formato: **WebP** (calidad 80). Si la codificación WebP falla, se usa **JPEG** (calidad 85).
- Los archivos **SVG** no generan thumbnail (`thumbnail_url` es `null`).
- Nombre del thumbnail: `{nombre_original}_thumb.webp` (o `_thumb.jpg` si falló WebP).
- Al eliminar una imagen se borran tanto el archivo original como el thumbnail.
- Los thumbnails se sirven por la misma ruta pública `/media/:name`.

**Ejemplo de recorte cover:**

Una imagen de 1200x800 se recorta al centro a 800x800 y se escala a 150x150:

```
1200x800 → center crop 800x800 → scale 150x150
```

---

## Upload Media

Sube un archivo de imagen al servidor. El archivo se almacena en disco (`MEDIA_DIR`), se genera un thumbnail y se registra en la base de datos.

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
  "thumbnail_url": "/media/1713500000000000000_thumb.webp",
  "created_at": "2026-04-19T12:00:00Z"
}
```

> Para SVGs, `thumbnail_url` es `null`.

**Errors:**
| Status | When |
|--------|------|
| 400 | No se envió archivo, MIME type no permitido |
| 401 | Token missing, invalid or expired |

---

## List Media

Lista los archivos multimedia paginados con cursor-based pagination, ordenados por ID ascendente.

```
GET /api/media
```

**Query Params:**
| Param | Type | Default | Max | Description |
|-------|------|---------|-----|-------------|
| `cursor` | int64 | - | - | ID del último elemento de la página anterior. Omite para obtener la primera página |
| `limit` | int32 | - | 100 | Cantidad de elementos por página. **Opcional**: si se omite, retorna todos los registros |
| `filter[columna]` | string | - | - | Filtro por columna. Ver [Filtros](#filtros-media) |

**Columnas filtrables:** `id`, `filename`, `mime_type`

**Response 200:**
```json
{
  "data": [
    {
      "id": 1,
      "filename": "foto.jpg",
      "mime_type": "image/jpeg",
      "size": 204800,
      "url": "/media/1713500000000000000.jpg",
      "thumbnail_url": "/media/1713500000000000000_thumb.webp",
      "created_at": "2026-04-19T12:00:00Z"
    },
    {
      "id": 2,
      "filename": "logo.svg",
      "mime_type": "image/svg+xml",
      "size": 2048,
      "url": "/media/1713500000000000001.svg",
      "thumbnail_url": null,
      "created_at": "2026-04-19T12:01:00Z"
    }
  ],
  "next_cursor": 2,
  "total": 2,
  "has_more": false
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Cursor o limit inválidos, columna de filtro desconocida |
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
  "thumbnail_url": "/media/1713500000000000000_thumb.webp",
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

Elimina un archivo multimedia. Borra tanto el registro de la base de datos como el fichero original y su thumbnail en disco.

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

Sirve un archivo multimedia desde disco. No requiere autenticación. Sirve tanto imágenes originales como thumbnails.

```
GET /media/:name
```

**Response 200:** El archivo binario con el `Content-Type` correspondiente.

**Errors:**
| Status | When |
|--------|------|
| 400 | Nombre vacío |
| 404 | Archivo no encontrado en disco |
