import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Use fixed test user credentials for login
TEST_USER_EMAIL = "testuser@example.com"
TEST_USER_PASSWORD = "TestPassw0rd!"


def login(email: str, password: str) -> str:
    url = f"{BASE_URL}/api/auth/login"
    payload = {"email": email, "password": password}
    resp = requests.post(url, json=payload, timeout=TIMEOUT)
    assert resp.status_code == 200, f"Login failed: {resp.text}"
    data = resp.json()
    token = data.get("token")
    assert token, "No token in login response"
    return token


def create_site(token: str, name: str, domain: str, options: dict, locales: list) -> dict:
    url = f"{BASE_URL}/api/sites"
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "name": name,
        "domain": domain,
        "options": options,
        "locales": locales,
    }
    resp = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    assert resp.status_code == 201, f"Site creation failed: {resp.text}"
    return resp.json()


def delete_site(token: str, site_id: int):
    url = f"{BASE_URL}/api/sites/{site_id}"
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.delete(url, headers=headers, timeout=TIMEOUT)
    assert resp.status_code == 204, f"Site deletion failed: {resp.text}"


def put_update_site(token: str, site_id: int, locale: str, data: dict) -> dict:
    url = f"{BASE_URL}/api/sites/{site_id}?locale={locale}"
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.put(url, json=data, headers=headers, timeout=TIMEOUT)
    assert resp.status_code == 200, f"Update site failed: {resp.text}"
    return resp.json()


def test_put_api_sites_id_update_site():
    token = login(TEST_USER_EMAIL, TEST_USER_PASSWORD)

    # Create a new site for update test
    unique_suffix = str(uuid.uuid4())[:8]
    original_name = f"Test Site {unique_suffix}"
    original_domain = f"test{unique_suffix}.example.com"
    original_options = {"theme": "light", "features": {"blog": True}}
    original_locales = [{"code": "en", "is_default": True}]

    site = create_site(token, original_name, original_domain, original_options, original_locales)
    site_id = site.get("id")
    assert site_id is not None, "Created site has no id"

    try:
        # Prepare updated data for PUT request
        updated_name = f"Updated Site {unique_suffix}"
        updated_domain = f"updated{unique_suffix}.example.com"
        updated_options = {"theme": "dark", "features": {"blog": False, "shop": True}}
        locale = "en"
        update_payload = {
            "name": updated_name,
            "domain": updated_domain,
            "options": updated_options,
        }

        updated_site = put_update_site(token, site_id, locale, update_payload)

        # Validate updated fields
        assert updated_site.get("name") == updated_name, "Site name not updated"
        assert updated_site.get("domain") == updated_domain, "Site domain not updated"
        assert "options" in updated_site and updated_site["options"] == updated_options, "Site options not updated"
        # Optionally validate locale if present
        if "locales" in updated_site:
            locale_codes = [loc.get("code") for loc in updated_site["locales"] if "code" in loc]
            assert locale in locale_codes, f"Locale '{locale}' missing in updated site locales"

    finally:
        # Clean up by deleting the created site
        delete_site(token, site_id)


test_put_api_sites_id_update_site()