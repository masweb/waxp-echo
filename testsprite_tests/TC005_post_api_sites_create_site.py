import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def get_auth_token(email: str, password: str) -> str:
    login_url = f"{BASE_URL}/api/auth/login"
    resp = requests.post(login_url, json={"email": email, "password": password}, timeout=TIMEOUT)
    resp.raise_for_status()
    data = resp.json()
    token = data.get("token")
    if not token:
        raise Exception("No token in login response")
    return token

def create_user(email: str, password: str):
    register_url = f"{BASE_URL}/api/auth/register"
    resp = requests.post(register_url, json={"email": email, "password": password}, timeout=TIMEOUT)
    if resp.status_code == 201:
        return
    elif resp.status_code == 409:
        # User already exists, ignore
        return
    else:
        resp.raise_for_status()

def test_post_api_sites_create_site():
    # Setup: Register and login user to get token
    test_email = f"test_user_{uuid.uuid4().hex[:8]}@example.com"
    test_password = "Password123!"
    create_user(test_email, test_password)
    token = get_auth_token(test_email, test_password)
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}

    # Generate unique domain and site data
    unique_domain = f"testdomain-{uuid.uuid4().hex[:8]}.com"
    site_payload = {
        "name": "Test Site",
        "domain": unique_domain,
        "options": {"theme": "default", "features": {"blog": True}},
        "locales": [
            {"code": "en", "is_default": True},
            {"code": "fr", "is_default": False}
        ]
    }

    site_url = f"{BASE_URL}/api/sites"
    created_site = None

    try:
        # POST /api/sites to create a new site
        response = requests.post(site_url, json=site_payload, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 201, f"Expected 201, got {response.status_code}. Response: {response.text}"
        created_site = response.json()

        # Validate response content
        assert "id" in created_site, "Created site missing 'id'"
        assert created_site.get("name") == site_payload["name"]
        assert created_site.get("domain") == site_payload["domain"]

        locales = created_site.get("locales")
        assert isinstance(locales, list) and len(locales) >= 1, "Locales missing or empty in created site"
        locale_codes = [loc.get("code") for loc in locales]
        for expected_locale in site_payload["locales"]:
            assert expected_locale["code"] in locale_codes

        # Check that default locale is present
        default_locales = [loc for loc in locales if loc.get("is_default") is True]
        assert len(default_locales) == 1, "Exactly one default locale expected"
        assert default_locales[0].get("code") == "en", "Default locale code mismatch"

        # Check that default pages or locale-aware pages created - typical indicator: "pages" field
        pages = created_site.get("pages")
        assert pages is None or isinstance(pages, list), "Pages should be list or None"
        if pages is not None:
            assert len(pages) >= 1, "Default pages expected to be auto-created"

    finally:
        # Cleanup: Delete the created site if it exists
        if created_site and "id" in created_site:
            delete_url = f"{BASE_URL}/api/sites/{created_site['id']}"
            del_resp = requests.delete(delete_url, headers=headers, timeout=TIMEOUT)
            # The specification says 204 expected on delete
            assert del_resp.status_code == 204 or del_resp.status_code == 404, f"Unexpected delete status: {del_resp.status_code}"

test_post_api_sites_create_site()
