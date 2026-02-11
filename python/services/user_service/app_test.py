from fastapi.testclient import TestClient

from user_service.app import app

client = TestClient(app)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json()["status"] == "ok"
    assert resp.json()["service"] == "user-service"


def test_create_and_list_users():
    resp = client.post("/users", json={"name": "Alice", "email": "alice@example.com"})
    assert resp.status_code == 201
    user = resp.json()
    assert user["name"] == "Alice"

    resp = client.get("/users")
    assert resp.status_code == 200
    assert len(resp.json()) >= 1


def test_get_user_not_found():
    resp = client.get("/users/nonexistent")
    assert resp.status_code == 404
