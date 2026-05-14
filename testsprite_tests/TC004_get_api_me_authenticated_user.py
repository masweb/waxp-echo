import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30


def test_get_api_me_authenticated_user():
    register_url = f"{BASE_URL}/api/auth/register"
    me_url = f"{BASE_URL}/api/me"

    # Generate a unique email for registration
    unique_email = f"testuser_{uuid.uuid4().hex}@example.com"
    password = "Password123!"

    headers = {"Content-Type": "application/json"}
    register_payload = {
        "email": unique_email,
        "password": password
    }

    # Register a new user to obtain a valid JWT token
    token = None
    try:
        reg_resp = requests.post(register_url, json=register_payload, headers=headers, timeout=TIMEOUT)
        assert reg_resp.status_code == 201, f"Registration failed with status {reg_resp.status_code}: {reg_resp.text}"
        reg_json = reg_resp.json()
        assert "token" in reg_json and "user" in reg_json, "Registration response missing token or user"
        token = reg_json["token"]
        user = reg_json["user"]
        assert "email" in user and user["email"] == unique_email, "Registered user email mismatch"

        # Call GET /api/me with the valid Bearer token
        me_headers = {"Authorization": f"Bearer {token}"}
        me_resp = requests.get(me_url, headers=me_headers, timeout=TIMEOUT)
        assert me_resp.status_code == 200, f"GET /api/me failed with status {me_resp.status_code}: {me_resp.text}"
        me_json = me_resp.json()
        assert "id" in me_json and "email" in me_json, "User profile missing id or email"
        assert me_json["email"] == unique_email, "Authenticated user email mismatch"

    finally:
        # No explicit delete user endpoint in PRD, so skipping resource cleanup.
        pass


test_get_api_me_authenticated_user()