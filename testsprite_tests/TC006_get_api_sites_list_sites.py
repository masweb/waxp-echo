import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Credentials for test user (must exist in the system)
TEST_USER_EMAIL = "testuser@example.com"
TEST_USER_PASSWORD = "testpassword123"

def get_auth_token():
    login_url = f"{BASE_URL}/api/auth/login"
    login_data = {
        "email": TEST_USER_EMAIL,
        "password": TEST_USER_PASSWORD
    }
    try:
        response = requests.post(login_url, json=login_data, timeout=TIMEOUT)
        assert response.status_code == 200, f"Login failed: {response.status_code}, {response.text}"
        token = response.json().get("token")
        assert token, "No token in login response"
        return token
    except requests.RequestException as e:
        raise AssertionError(f"Login request failed: {str(e)}")

def create_site(auth_token):
    create_url = f"{BASE_URL}/api/sites"
    unique_suffix = str(uuid.uuid4())
    site_data = {
        "name": "Test Site " + unique_suffix,
        "domain": "testsite-" + unique_suffix + ".com",
        "options": {},
        "locales": [
            {"code": "en", "is_default": True}
        ]
    }
    headers = {
        "Authorization": f"Bearer {auth_token}"
    }
    try:
        response = requests.post(create_url, json=site_data, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 201, f"Site creation failed: {response.status_code}, {response.text}"
        site = response.json()
        assert "id" in site, "Created site has no id"
        return site["id"]
    except requests.RequestException as e:
        raise AssertionError(f"Create site request failed: {str(e)}")

def delete_site(auth_token, site_id):
    delete_url = f"{BASE_URL}/api/sites/{site_id}"
    headers = {
        "Authorization": f"Bearer {auth_token}"
    }
    try:
        response = requests.delete(delete_url, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 204 or response.status_code == 404, (
            f"Delete site failed: {response.status_code}, {response.text}"
        )
    except requests.RequestException as e:
        raise AssertionError(f"Delete site request failed: {str(e)}")

def test_get_api_sites_list_sites():
    auth_token = get_auth_token()
    headers = {"Authorization": f"Bearer {auth_token}"}
    site_id = None
    try:
        # Create a new site to ensure there is at least one site to retrieve/filter
        site_id = create_site(auth_token)

        # 1. Test GET /api/sites without filters (default pagination)
        url = f"{BASE_URL}/api/sites"
        response = requests.get(url, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 200, f"GET /api/sites failed: {response.status_code}, {response.text}"
        data = response.json()
        assert isinstance(data, dict), "Response is not a JSON object"
        assert "items" in data or "sites" in data, "Response missing 'items' or 'sites' list"
        # Accept either 'items' or 'sites' as per typical pagination key
        sites_list = data.get("items") or data.get("sites") or []
        assert isinstance(sites_list, list), "Sites list is not a list"
        # Check at least one site present (the created one)
        assert any(str(site.get("id")) == str(site_id) for site in sites_list), "Created site not in list"

        # 2. Test GET /api/sites with filter[name] to filter list
        filter_name = "Test Site"
        params = {
            "limit": 10,
            "filter[name]": filter_name
        }
        response = requests.get(url, headers=headers, params=params, timeout=TIMEOUT)
        assert response.status_code == 200, f"GET /api/sites with filter failed: {response.status_code}, {response.text}"
        data_filtered = response.json()
        filtered_list = data_filtered.get("items") or data_filtered.get("sites") or []
        assert isinstance(filtered_list, list), "Filtered sites list is not a list"
        # All returned sites should have name containing filter_name substring (case-insensitive)
        for site in filtered_list:
            name = site.get("name", "")
            assert filter_name.lower() in name.lower(), f"Site name {name} does not contain filter {filter_name}"

    finally:
        if site_id:
            delete_site(auth_token, site_id)

test_get_api_sites_list_sites()