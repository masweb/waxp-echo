# API Reference

## Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/auth/register` | Register new user |
| `POST` | `/api/auth/login` | Login user |
| `GET` | `/api/me` | Get current user (requires JWT) |

## Sites

| Method | Endpoint | Query Params | Description |
|--------|----------|-------------|-------------|
| `GET` | `/api/sites` | `cursor`, `limit`, `filter[*]` | List sites (paginated) |
| `POST` | `/api/sites` | | Create site with default pages |
| `GET` | `/api/sites/:id` | `locale` (opcional) | Get site (options resolved to locale, default if omitted) |
| `PUT` | `/api/sites/:id` | `locale` **(required)** | Update site (options merged by locale) |
| `DELETE` | `/api/sites/:id` | | Delete site |
| `POST` | `/api/sites/:id/locales` | | Add locale to site |
| `DELETE` | `/api/sites/:id/locales/:localeCode` | | Remove locale from site |

## Pages

| Method | Endpoint | Query Params | Description |
|--------|----------|-------------|-------------|
| `GET` | `/api/sites/:id/pages` | `cursor`, `limit`, `filter[*]` | List pages (paginated, filterable) |
| `POST` | `/api/sites/:id/pages` | `locale` **(required)** | Create page or post |
| `GET` | `/api/sites/:id/pages/:pageId` | `locale` **(required)** | Get page (layout resolved to locale) |
| `PUT` | `/api/sites/:id/pages/:pageId` | `locale` **(required)** | Update page (layout merged by locale) |
| `DELETE` | `/api/sites/:id/pages/:pageId` | | Delete page |
| `GET` | `/api/sites/:id/routes` | | Get all routes (for vue-router) |
| `POST` | `/api/sites/:id/sections/next-id` | | Get next unique section ID |
| `POST` | `/api/sites/:id/blocks/next-id` | | Get next unique block ID |

## Media

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/media` | Upload media file |
| `GET` | `/api/media` | List media (paginated) |
| `GET` | `/api/media/:id` | Get media |
| `DELETE` | `/api/media/:id` | Delete media |
| `GET` | `/media/:name` | Serve media file (public) |

---

## Authentication

Todas las rutas de Sites, Pages, Locales y Media requieren autenticación JWT:

```
Authorization: Bearer <token>
```

## Paginación

Ver [pagination.md](./pagination.md) para detalle de paginación cursor-based y filtros.

## Filtros disponibles

Ver [pagination.md](./pagination.md) para operadores soportados (`_like`, `_eq`, `_neq`, `_gt`, `_gte`, `_lt`, `_lte`, `_in`, `_isnull`).
