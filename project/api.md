# API

## Authentication

### Register

```
POST /api/auth/register
```

**Body:**
```json
{
  "email": "user@example.com",
  "password": "12345678"
}
```

**Response 201:**
```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Email or password empty, or password < 8 chars |
| 409 | Email already exists |

---

### Login

```
POST /api/auth/login
```

**Body:**
```json
{
  "email": "user@example.com",
  "password": "12345678"
}
```

**Response 200:**
```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Email or password empty |
| 401 | Invalid credentials |

---

### Me

```
GET /api/me
```

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "id": 1,
  "email": "user@example.com"
}
```

**Errors:**
| Status | When |
|--------|------|
| 401 | Token missing, invalid or expired |

---

### Health

```
GET /health
```

**Response 200:**
```json
{
  "status": "ok"
}
```
