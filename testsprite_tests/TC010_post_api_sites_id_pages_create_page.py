import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Test user credentials
AUTH_EMAIL = "testuser_tc010@example.com"
AUTH_PASSWORD = "StrongPass123"

def register_user(email, password):
    reg_url = f"{BASE_URL}/api/auth/register"
    resp = requests.post(reg_url, json={"email": email, "password": password}, timeout=TIMEOUT)
    if resp.status_code == 201:
        return
    elif resp.status_code == 409:
        # user already exists
        return
    else:
        assert False, f"Registration failed with status {resp.status_code}"

def get_auth_token(email, password):
    # Ensure user is registered before login
    register_user(email, password)

    login_url = f"{BASE_URL}/api/auth/login"
    resp = requests.post(login_url, json={"email": email, "password": password}, timeout=TIMEOUT)
    assert resp.status_code == 200, f"Login failed with status {resp.status_code}"
    data = resp.json()
    token = data.get("token")
    assert token and isinstance(token, str), "No valid token in login response"
    return token

def create_site(auth_token):
    site_url = f"{BASE_URL}/api/sites"
    unique_domain = f"tc010-{uuid.uuid4().hex[:8]}.example.com"
    site_data = {
        "name": "Test Site TC010",
        "domain": unique_domain,
        "options": {},
        "locales": [{"code": "en", "is_default": True}]
    }
    headers = {"Authorization": f"Bearer {auth_token}"}
    resp = requests.post(site_url, json=site_data, headers=headers, timeout=TIMEOUT)
    assert resp.status_code == 201, f"Site creation failed with status {resp.status_code}"
    site = resp.json()
    site_id = site.get("id")
    assert site_id is not None, "Created site response missing id"
    return site_id

def delete_site(auth_token, site_id):
    del_url = f"{BASE_URL}/api/sites/{site_id}"
    headers = {"Authorization": f"Bearer {auth_token}"}
    resp = requests.delete(del_url, headers=headers, timeout=TIMEOUT)
    assert resp.status_code == 204, f"Site deletion failed with status {resp.status_code}"

def test_post_api_sites_id_pages_create_page():
    # Obtain auth token
    auth_token = get_auth_token(AUTH_EMAIL, AUTH_PASSWORD)
    headers = {
        "Authorization": f"Bearer {auth_token}",
        "Content-Type": "application/json"
    }
    site_id = None
    page_id = None
    try:
        # Create site to POST page against
        site_id = create_site(auth_token)

        url = f"{BASE_URL}/api/sites/{site_id}/pages"
        params = {"locale": "en"}

        page_payload = {
            "type": "page",
            "slugs": ["test-page-tc010"],
            "layout": {
                "sections": [
                    {"id": 1, "type": "header", "content": {"text": "Header Section"}},
                    {"id": 2, "type": "body", "content": {"text": "Body Section"}},
                    {"id": 3, "type": "footer", "content": {"text": "Footer Section"}},
                    {"id": 4, "type": "sidebar", "content": {"text": "Sidebar Section"}}
                ]
            }
        }

        resp = requests.post(url, headers=headers, params=params, json=page_payload, timeout=TIMEOUT)
        assert resp.status_code == 201, f"Page creation failed with status {resp.status_code}"
        data = resp.json()

        # Validate page fields
        page_id = data.get("id")
        assert page_id is not None, "Response missing page id"
        assert data.get("type") == "page", "Page type mismatch"
        assert "layout" in data, "Response missing layout"
        layout = data["layout"]
        # Check that layout has "sections" with expected count (4)
        sections = layout.get("sections")
        assert sections is not None and isinstance(sections, list), "Layout sections missing or invalid"
        assert len(sections) == 4, f"Expected 4 layout sections, got {len(sections)}"
        # Optional: check that slugs in response match input
        slugs = data.get("slugs")
        assert slugs == ["test-page-tc010"], "Slugs mismatch in response"
    finally:
        # Cleanup: Delete created page and site if exist
        if page_id and site_id:
            del_page_url = f"{BASE_URL}/api/sites/{site_id}/pages/{page_id}"
            try:
                resp = requests.delete(del_page_url, headers={"Authorization": f"Bearer {auth_token}"}, timeout=TIMEOUT)
                # 204 expected, ignore otherwise
            except Exception:
                pass
        if site_id:
            try:
                delete_site(auth_token, site_id)
            except Exception:
                pass

test_post_api_sites_id_pages_create_page()
