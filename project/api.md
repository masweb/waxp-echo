# API Reference

## Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/auth/register` | Register new user |
| `POST` | `/api/auth/login` | Login user |
| `GET` | `/api/me` | Get current user (requires JWT) |

## Sites

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/sites` | List sites (paginated) |
| `POST` | `/api/sites` | Create site |
| `POST` | `/api/sites/init` | Create site with default pages |
| `GET` | `/api/sites/:id` | Get site |
| `PUT` | `/api/sites/:id` | Update site |
| `DELETE` | `/api/sites/:id` | Delete site |
| `POST` | `/api/sites/:id/locales` | Add locale to site |
| `DELETE` | `/api/sites/:id/locales/:localeCode` | Remove locale from site |

## Pages

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/sites/:id/pages` | List pages (paginated, filterable) |
| `POST` | `/api/sites/:id/pages` | Create page or post |
| `GET` | `/api/sites/:id/pages/:pageId` | Get page |
| `PUT` | `/api/sites/:id/pages/:pageId` | Update page |
| `DELETE` | `/api/sites/:id/pages/:pageId` | Delete page |
| `GET` | `/api/sites/:id/routes` | Get all routes (for vue-router) |

---

## Authentication

Todas las rutas de Sites, Pages y Locales requieren autenticación JWT:

```
Authorization: Bearer <token>
```

## Paginación

Ver [pagination.md](./pagination.md) para detalle de paginación cursor-based y filtros.

## Filtros disponibles

Ver [pagination.md](./pagination.md) para operadores soportados (`_like`, `_eq`, `_neq`, `_gt`, `_gte`, `_lt`, `_lte`, `_in`, `_isnull`).
