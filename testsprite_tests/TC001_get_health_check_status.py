import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_get_health_check_status():
    url = f"{BASE_URL}/health"
    try:
        response = requests.get(url, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request to {url} failed with exception: {e}"
    assert response.status_code == 200, f"Expected status code 200 but got {response.status_code}"
    try:
        json_data = response.json()
    except ValueError:
        assert False, "Response is not valid JSON"
    assert json_data == {"status": "ok"}, f"Expected response body {{'status':'ok'}} but got {json_data}"

test_get_health_check_status()