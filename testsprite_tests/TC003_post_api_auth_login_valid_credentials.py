import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_api_auth_login_valid_credentials():
    # Given valid credentials (these should exist in the system for testing)
    email = "testuser@example.com"
    password = "ValidPass123!"

    url = f"{BASE_URL}/api/auth/login"
    payload = {
        "email": email,
        "password": password
    }
    headers = {
        "Content-Type": "application/json"
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    assert response.status_code == 200, f"Expected status code 200 but got {response.status_code}"
    try:
        data = response.json()
    except ValueError:
        assert False, "Response is not valid JSON"

    assert "token" in data, "Response JSON does not contain 'token'"
    assert isinstance(data["token"], str) and len(data["token"]) > 0, "'token' is not a non-empty string"
    assert "user" in data, "Response JSON does not contain 'user'"
    user = data["user"]
    assert isinstance(user, dict), "'user' is not a dict"
    assert "id" in user and isinstance(user["id"], int), "'user.id' missing or not int"
    assert "email" in user and user["email"] == email, "'user.email' missing or does not match login email"

test_post_api_auth_login_valid_credentials()