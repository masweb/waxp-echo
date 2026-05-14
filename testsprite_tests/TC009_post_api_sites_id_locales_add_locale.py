import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Credentials for authentication - must be valid in the system or register a new user
TEST_USER_EMAIL = "testuser_locale_add@example.com"
TEST_USER_PASSWORD = "StrongPassw0rd!"

def get_auth_token():
    # Try to register, if already exists then login
    register_payload = {
        "email": TEST_USER_EMAIL,
        "password": TEST_USER_PASSWORD
    }
    try:
        r = requests.post(f"{BASE_URL}/api/auth/register", json=register_payload, timeout=TIMEOUT)
        if r.status_code == 201:
            return r.json()["token"]
    except requests.RequestException:
        pass
    # Login fallback
    login_payload = {
        "email": TEST_USER_EMAIL,
        "password": TEST_USER_PASSWORD
    }
    r = requests.post(f"{BASE_URL}/api/auth/login", json=login_payload, timeout=TIMEOUT)
    r.raise_for_status()
    return r.json()["token"]

def create_site(auth_token):
    unique_suffix = str(uuid.uuid4())[:8]
    site_payload = {
        "name": f"Test Site {unique_suffix}",
        "domain": f"testsite-{unique_suffix}.example.com",
        "options": {},
        "locales": [
            {"code": "en", "is_default": True}
        ]
    }
    headers = {
        "Authorization": f"Bearer {auth_token}"
    }
    r = requests.post(f"{BASE_URL}/api/sites", json=site_payload, headers=headers, timeout=TIMEOUT)
    r.raise_for_status()
    assert r.status_code == 201
    return r.json()  # Return the created site object

def delete_site(auth_token, site_id):
    headers = {
        "Authorization": f"Bearer {auth_token}"
    }
    r = requests.delete(f"{BASE_URL}/api/sites/{site_id}", headers=headers, timeout=TIMEOUT)
    if r.status_code not in (204, 404):
        r.raise_for_status()

def test_post_api_sites_id_locales_add_locale():
    auth_token = get_auth_token()
    headers = {
        "Authorization": f"Bearer {auth_token}",
        "Content-Type": "application/json"
    }

    site = create_site(auth_token)
    site_id = site["id"]

    new_locale_payload = {
        "code": "ca",
        "is_default": False
    }

    try:
        # Add locale "ca"
        r = requests.post(
            f"{BASE_URL}/api/sites/{site_id}/locales",
            json=new_locale_payload,
            headers=headers,
            timeout=TIMEOUT
        )
        assert r.status_code == 201, f"Expected 201 Created, got {r.status_code}"
        locale = r.json()
        assert locale.get("code") == new_locale_payload["code"], "Locale code mismatch"
        assert locale.get("is_default") == new_locale_payload["is_default"], "Locale is_default mismatch"
    finally:
        # Cleanup: delete site after test
        delete_site(auth_token, site_id)

test_post_api_sites_id_locales_add_locale()