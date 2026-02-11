from fastapi.testclient import TestClient

from product_service.app import app

client = TestClient(app)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json()["status"] == "ok"
    assert resp.json()["service"] == "product-service"


def test_create_and_list_products():
    resp = client.post(
        "/products",
        json={"name": "Widget", "price": 9.99, "description": "A fine widget"},
    )
    assert resp.status_code == 201
    product = resp.json()
    assert product["name"] == "Widget"
    assert product["price"] == 9.99

    resp = client.get("/products")
    assert resp.status_code == 200
    assert len(resp.json()) >= 1


def test_get_product_not_found():
    resp = client.get("/products/nonexistent")
    assert resp.status_code == 404
