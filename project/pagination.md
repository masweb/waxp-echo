# Paginación

Todos los endpoints de listado usan **paginación cursor-based** por ID ascendente con soporte para **filtros**.

## Parámetros

| Param | Type | Default | Max | Description |
|-------|------|---------|-----|-------------|
| `cursor` | int64 | - | - | ID del último elemento de la página anterior. Omite para obtener la primera página |
| `limit` | int32 | - | 100 | Cantidad de elementos por página. **Opcional**: si se omite, retorna todos los registros |
| `filter[columna]` | string | - | - | Filtro por columna. Ver [Filtros](#filtros) |

## Filtros

Los filtros se envían como query params con el prefijo `filter[...]`.

### Formato

```
GET /api/sites?filter[columna]=valor
GET /api/sites?filter[columna_operador]=valor
```

### Operadores

Se añaden como sufijo al nombre de la columna separados por `_`:

| Operador | Sufijo | SQL | Ejemplo | Descripción |
|----------|--------|-----|---------|-------------|
| Igualdad | (ninguno) | `=` | `filter[name]=Mi Blog` | Coincidencia exacta |
| Desigualdad | `_neq` / `_not` | `!=` | `filter[domain_not]=old.com` | Diferente de |
| Contiene | `_like` | `ILIKE` | `filter[name_like]=blog` | Contiene el texto (case-insensitive) |
| Mayor que | `_gt` | `>` | `filter[id_gt]=5` | Mayor que |
| Mayor o igual | `_gte` | `>=` | `filter[id_gte]=5` | Mayor o igual que |
| Menor que | `_lt` | `<` | `filter[id_lt]=10` | Menor que |
| Menor o igual | `_lte` | `<=` | `filter[id_lte]=10` | Menor o igual que |
| En lista | `_in` | `IN` | `filter[id_in]=1,2,3` | La columna está en la lista (valores separados por coma) |
| Es nulo | `_isnull` | `IS NULL` | `filter[name_isnull]=true` | `true` = IS NULL, `false` = IS NOT NULL |

### Combinar filtros

Múltiples filtros se combinan con `AND`:

```
GET /api/sites?filter[name_like]=blog&filter[id_gt]=5
```

Genera: `WHERE name ILIKE '%blog%' AND id > 5`

### Combinar con paginación

Filtros y paginación funcionan juntos. El `cursor` se aplica sobre los resultados filtrados:

```
GET /api/sites?filter[name_like]=blog&limit=10&cursor=20
```

Genera: `WHERE id > 20 AND name ILIKE '%blog%' ORDER BY id ASC LIMIT 10`

El `total` en la respuesta refleja la cantidad de resultados filtrados, no el total de la tabla.

### Columnas disponibles

Cada endpoint define qué columnas se pueden filtrar. Columnas no listadas son ignoradas.

**Sites:** `id`, `name`, `domain`

## Respuesta

Los endpoints paginados retornan siempre esta estructura:

```json
{
  "data": [...],
  "next_cursor": 42,
  "total": 150,
  "has_more": true
}
```

| Campo | Type | Description |
|-------|------|-------------|
| `data` | array | Elementos de la página actual |
| `next_cursor` | int64 \| null | ID del último elemento. Usar como `cursor` en la siguiente petición. `null` si no hay más resultados |
| `total` | int64 | Total de elementos que coinciden con los filtros |
| `has_more` | bool | `true` si existen más elementos después de esta página |

## Ejemplos

### Sin paginación ni filtros (traer todos)
```
GET /api/sites
```

### Con filtros, sin paginación
```
GET /api/sites?filter[name_like]=blog
```

### Primera página filtrada
```
GET /api/sites?filter[name_like]=blog&limit=10
```

### Páginas siguientes filtradas
```
GET /api/sites?filter[name_like]=blog&limit=10&cursor=42
```

### Múltiples filtros
```
GET /api/sites?filter[name_like]=blog&filter[id_gte]=5&filter[id_lte]=50
```

## Comportamiento

- Los resultados se ordenan por `id ASC` (orden de inserción).
- El cursor es el ID del último registro recibido; la siguiente consulta usa `WHERE id > cursor AND [filtros]`.
- Si `limit` es mayor a 100 se clampa automáticamente a 100.
- Si no se envía `limit`, se retornan **todos** los registros filtrados (sin paginación).
- Si `data` está vacío, se retorna `[]` (nunca `null`).
- Los valores de filtro no reconocidos son ignorados silenciosamente.

## Implementación

**Filtros:** `internal/filter/filter.go` — paquete `filter` con `Builder` genérico reutilizable en todos los modelos.

```go
builder := filter.NewBuilder(map[string]string{
    "name":   "name",
    "domain": "domain",
    "id":     "id",
})
builder.Parse(c.Request().URL.Query())
result := builder.Build(cursor)
// result.WhereClause → " WHERE name ILIKE $1 AND id > $2"
// result.Args → []any{"%blog%", int64(5)}
```

**Paginación:** tipo genérico `PaginatedResponse[T]` definido en `internal/handler/site.go`:

```go
type PaginatedResponse[T any] struct {
    Data       []T    `json:"data"`
    NextCursor *int64 `json:"next_cursor"`
    Total      int64  `json:"total"`
    HasMore    bool   `json:"has_more"`
}
```


Para reutilizar en otros modelos, solo hay que:

Definir el map[string]string de columnas permitidas
Llamar a builder.Parse() + builder.Build(cursor)
Usar result.WhereClause y result.Args en las queries
