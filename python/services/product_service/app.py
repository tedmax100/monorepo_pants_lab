"""Product Service — FastAPI application for managing products."""

from __future__ import annotations

from datetime import datetime
from uuid import uuid4

from fastapi import FastAPI, HTTPException

from common.models import HealthResponse, Product, ProductBase

app = FastAPI(title="Product Service", version="0.2.0")

# In-memory store
_products: dict[str, Product] = {}


@app.get("/health", response_model=HealthResponse)
def health():
    return HealthResponse(status="ok", service="product-service")


@app.get("/products", response_model=list[Product])
def list_products():
    return list(_products.values())


@app.get("/products/{product_id}", response_model=Product)
def get_product(product_id: str):
    if product_id not in _products:
        raise HTTPException(status_code=404, detail="Product not found")
    return _products[product_id]


@app.post("/products", response_model=Product, status_code=201)
def create_product(payload: ProductBase):
    product = Product(
        id=str(uuid4()),
        name=payload.name,
        price=payload.price,
        description=payload.description,
        created_at=datetime.now(),
    )
    _products[product.id] = product
    return product
