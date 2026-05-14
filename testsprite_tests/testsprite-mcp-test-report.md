# TestSprite AI Testing Report(MCP)

---

## 1️⃣ Document Metadata
- **Project Name:** echo
- **Date:** 2026-05-11
- **Prepared by:** TestSprite AI Team
- **Environment:** Backend (Go / Echo v5 / PostgreSQL)
- **Server:** http://localhost:8080
- **Total Tests:** 10
- **Passed:** 5 (50%)
- **Failed:** 5 (50%)

---

## 2️⃣ Requirement Validation Summary

### Requirement: Health Check

| Test ID | Title | Status |
|---------|-------|--------|
| TC001 | get_health_check_status | ✅ Passed |

#### TC001 — get_health_check_status
- **Status:** ✅ Passed
- **Endpoint:** `GET /health`
- **Analysis:** El endpoint responde correctamente con `{"status":"ok"}` y status 200. Funciona sin autenticación como se espera.

---

### Requirement: Authentication

| Test ID | Title | Status |
|---------|-------|--------|
| TC002 | post_api_auth_register_new_user | ✅ Passed |
| TC003 | post_api_auth_login_valid_credentials | ❌ Failed |
| TC004 | get_api_me_authenticated_user | ✅ Passed |

#### TC002 — post_api_auth_register_new_user
- **Status:** ✅ Passed
- **Endpoint:** `POST /api/auth/register`
- **Analysis:** El registro de usuarios funciona correctamente. Genera un email único con UUID, envía credenciales válidas y recibe status 201 con token JWT y objeto user. El test limpia adecuadamente.

#### TC003 — post_api_auth_login_valid_credentials
- **Status:** ❌ Failed
- **Endpoint:** `POST /api/auth/login`
- **Error:** `AssertionError: Expected status code 200 but got 401`
- **Analysis:** El test utiliza credenciales hardcodeadas (`testuser@example.com` / `ValidPass123!`) que no existen en la base de datos. A diferencia de TC002 que registra un usuario nuevo, este test asume que el usuario ya existe. Debería registrar un usuario primero y luego hacer login con esas credenciales, como hacen TC004 y TC005.

#### TC004 — get_api_me_authenticated_user
- **Status:** ✅ Passed
- **Endpoint:** `GET /api/me`
- **Analysis:** Funciona correctamente. Registra un usuario nuevo, obtiene el token del registro y lo usa para autenticarse en `/api/me`. Devuelve el perfil del usuario autenticado correctamente.

---

### Requirement: Sites Management

| Test ID | Title | Status |
|---------|-------|--------|
| TC005 | post_api_sites_create_site | ✅ Passed |
| TC006 | get_api_sites_list_sites | ❌ Failed |
| TC007 | put_api_sites_id_update_site | ❌ Failed |
| TC008 | put_api_sites_id_live_set_site_live | ❌ Failed |

#### TC005 — post_api_sites_create_site
- **Status:** ✅ Passed
- **Endpoint:** `POST /api/sites`
- **Analysis:** Creación de sites funciona correctamente. Registra un usuario, obtiene token, crea un site con dominio único y locales. Verifica que la respuesta incluye id, name, domain, locales y al menos un locale por defecto. Limpia eliminando el site al finalizar.

#### TC006 — get_api_sites_list_sites
- **Status:** ❌ Failed
- **Endpoint:** `GET /api/sites`
- **Error:** `AssertionError: Login failed: 401, {"error":"invalid credentials","code":401}`
- **Analysis:** Mismo problema que TC003: usa credenciales hardcodeadas (`testuser@example.com`) que no existen. El login falla y no puede ejecutar el test de listado. Además, el test espera la clave `items` o `sites` en la respuesta paginada, pero la API usa `data` según la documentación.

#### TC007 — put_api_sites_id_update_site
- **Status:** ❌ Failed
- **Endpoint:** `PUT /api/sites/:id`
- **Error:** `AssertionError: Login failed: {"error":"invalid credentials","code":401}`
- **Analysis:** Mismo problema de credenciales hardcodeadas. El test no puede progresar más allá del login.

#### TC008 — put_api_sites_id_live_set_site_live
- **Status:** ❌ Failed
- **Endpoint:** `PUT /api/sites/:id/live`
- **Error:** `AssertionError: Login failed: {"error":"invalid credentials","code":401}`
- **Analysis:** Mismo problema de credenciales hardcodeadas. No llega a ejecutar la lógica del test.

---

### Requirement: Locales Management

| Test ID | Title | Status |
|---------|-------|--------|
| TC009 | post_api_sites_id_locales_add_locale | ✅ Passed |

#### TC009 — post_api_sites_id_locales_add_locale
- **Status:** ✅ Passed
- **Endpoint:** `POST /api/sites/:id/locales`
- **Analysis:** Funciona correctamente. Registra un usuario nuevo, crea un site, añade un locale "ca" y verifica que la respuesta contiene el código y el flag is_default correctos. Limpia recursos al finalizar.

---

### Requirement: Pages Management

| Test ID | Title | Status |
|---------|-------|--------|
| TC010 | post_api_sites_id_pages_create_page | ❌ Failed |

#### TC010 — post_api_sites_id_pages_create_page
- **Status:** ❌ Failed
- **Endpoint:** `POST /api/sites/:id/pages`
- **Error:** `AssertionError: Page creation failed with status 400`
- **Analysis:** El test envía los slugs en formato incorrecto: `["test-page-tc010"]` (array de strings simples) cuando la API espera `[{"locale_code": "en", "slug": "test-page-tc010"}]` (array de objetos con locale_code y slug). También envía un layout con formato `{"sections": [...]}` que no coincide con la estructura esperada por la API (array directo de secciones con id, mobile, tablet, desktop, blocks). El status 400 indica que el servidor rechaza el body por formato inválido.

---

## 3️⃣ Coverage & Matching Metrics

- **50.00%** de tests pasaron (5 de 10)

| Requirement | Total Tests | ✅ Passed | ❌ Failed |
|-------------|-------------|-----------|-----------|
| Health Check | 1 | 1 | 0 |
| Authentication | 3 | 2 | 1 |
| Sites Management | 4 | 1 | 3 |
| Locales Management | 1 | 1 | 0 |
| Pages Management | 1 | 0 | 1 |

**Endpoints no cubiertos por los tests:**
- `DELETE /api/sites/:id` (solo se usa en cleanup, no tiene test dedicado)
- `GET /api/sites/:id`
- `DELETE /api/sites/:id/locales/:localeCode`
- `GET /api/sites/:id/pages/:pageId`
- `PUT /api/sites/:id/pages/:pageId`
- `DELETE /api/sites/:id/pages/:pageId`
- `GET /api/sites/:id/pages/:pageId/revisions`
- `POST /api/sites/:id/pages/:pageId/revisions/:revisionId/restore`
- `GET /api/sites/:id/routes`
- `POST /api/sites/:id/sections/next-id`
- `POST /api/sites/:id/blocks/next-id`
- `POST /api/media` (upload)
- `GET /api/media`
- `GET /api/media/:id`
- `DELETE /api/media/:id`
- `GET /media/:name` (public serve)

---

## 4️⃣ Key Gaps / Risks

### 1. Credenciales hardcodeadas en tests de TestSprite (4 tests afectados)
Los tests TC003, TC006, TC007 y TC008 usan credenciales fijas (`testuser@example.com`) que no existen en la base de datos. TestSprite genera estos tests asumiendo que hay un usuario preexistente. Los tests que sí pasan (TC004, TC005, TC009) registran un usuario nuevo antes de operar. **Mitigación:** Se podría crear un usuario semilla en la base de datos antes de ejecutar los tests, o bien modificar los tests para que sigan el patrón register-then-use.

### 2. Formato incorrecto del payload en TC010
El test de creación de páginas envía slugs como strings simples en vez de objetos `{locale_code, slug}`, y el layout tiene una estructura que no coincide con el schema de la API. Esto revela que TestSprite no infirió correctamente el schema del body a partir del PRD. **Mitigación:** Proporcionar ejemplos más explícitos en el code_summary.yaml o usar additional instructions con formato de ejemplo.

### 3. Baja cobertura de endpoints
Solo 10 endpoints de los ~20 disponibles están cubiertos por tests. Faltan especialmente los de Media (upload, list, delete), Pages (get, update, delete, revisions, routes), y los contadores de IDs (sections/next-id, blocks/next-id).

### 4. Cobertura de edge cases nula
No hay tests de validación de errores (400, 401, 404, 409), paginación con cursor, filtros combinados, ni edge cases como contraseñas cortas, emails duplicados, dominios duplicados, o slugs inválidos.

---
