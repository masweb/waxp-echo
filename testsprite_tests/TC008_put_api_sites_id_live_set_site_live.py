import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Credentials for login - these should exist in the system beforehand
TEST_USER_EMAIL = "testuser@example.com"
TEST_USER_PASSWORD = "TestPass123!"

def authenticate():
    """Authenticate and return JWT token."""
    login_url = f"{BASE_URL}/api/auth/login"
    login_payload = {
        "email": TEST_USER_EMAIL,
        "password": TEST_USER_PASSWORD
    }
    response = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
    assert response.status_code == 200, f"Login failed: {response.text}"
    data = response.json()
    token = data.get("token")
    assert token, "JWT token not found in login response"
    return token

def create_site(token):
    """Create a new site and return its site object."""
    url = f"{BASE_URL}/api/sites"
    unique_domain = f"testdomain-{uuid.uuid4()}.com"
    site_payload = {
        "name": "Test Site for Live Set",
        "domain": unique_domain,
        "options": {},
        "locales": [{"code": "en", "is_default": True}]
    }
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.post(url, json=site_payload, headers=headers, timeout=TIMEOUT)
    assert response.status_code == 201, f"Site creation failed: {response.text}"
    return response.json()

def delete_site(token, site_id):
    """Delete the specified site."""
    url = f"{BASE_URL}/api/sites/{site_id}"
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.delete(url, headers=headers, timeout=TIMEOUT)
    assert response.status_code == 204, f"Failed to delete site {site_id}: {response.text}"

def test_put_api_sites_id_live_set_site_live():
    token = authenticate()
    headers = {"Authorization": f"Bearer {token}"}
    site = create_site(token)
    site_id = site.get("id")
    assert site_id is not None, "Site id is missing from create site response"

    try:
        url = f"{BASE_URL}/api/sites/{site_id}/live"
        response = requests.put(url, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 200, f"PUT /api/sites/:id/live returned status {response.status_code}: {response.text}"
        site_response = response.json()
        # Validate response has the site object with an "id" and expected live status
        assert site_response.get("id") == site_id, "Returned site id does not match requested"
        # Check live status - based on PRD, assume "live" field should be True or active status included
        # We will check for "live" key existence and its truthiness if present
        if "live" in site_response:
            assert site_response["live"] in (True, "active", "yes", 1), "Site live status is not active"
        # Otherwise, we accept just that the site object is returned
        assert "name" in site_response, "Site object missing 'name' field"
        assert "domain" in site_response, "Site object missing 'domain' field"
    finally:
        delete_site(token, site_id)

test_put_api_sites_id_live_set_site_live()