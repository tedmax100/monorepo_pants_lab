from datetime import datetime

from common.models import HealthResponse, Product, User


def test_user_model():
    user = User(
        id="u-1",
        name="Alice",
        email="alice@example.com",
        created_at=datetime.now(),
    )
    assert user.name == "Alice"
    assert user.id == "u-1"


def test_product_model():
    product = Product(
        id="p-1",
        name="Widget",
        price=9.99,
        description="A fine widget",
        created_at=datetime.now(),
    )
    assert product.price == 9.99


def test_health_response():
    health = HealthResponse(status="ok", service="test")
    assert health.status == "ok"
