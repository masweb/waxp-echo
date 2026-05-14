import requests
import uuid

BASE_URL = "http://localhost:8080"

def test_post_api_auth_register_new_user():
    url = f"{BASE_URL}/api/auth/register"
    # Create a unique email to avoid conflicts
    unique_email = f"testuser_{uuid.uuid4()}@example.com"
    payload = {
        "email": unique_email,
        "password": "StrongPass123"
    }
    headers = {
        "Content-Type": "application/json"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
        # Assert status code 201 Created
        assert response.status_code == 201, f"Expected status 201, got {response.status_code}. Response: {response.text}"
        data = response.json()
        # Assert response contains token and user object with id and email
        assert "token" in data, "Response JSON missing 'token'"
        assert isinstance(data["token"], str) and len(data["token"]) > 0, "'token' is empty or not a string"
        assert "user" in data, "Response JSON missing 'user'"
        user = data["user"]
        assert isinstance(user, dict), "'user' is not a dictionary"
        assert "id" in user, "'user' object missing 'id'"
        assert isinstance(user["id"], int), "'user.id' is not an integer"
        assert "email" in user, "'user' object missing 'email'"
        assert user["email"] == unique_email, f"User email mismatch: expected {unique_email}, got {user['email']}"
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

test_post_api_auth_register_new_user()