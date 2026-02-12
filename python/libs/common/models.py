"""Shared Pydantic models used across Python services."""

from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, EmailStr


class UserBase(BaseModel):
    name: str
    email: str
    # nickname: str | None = None


class User(UserBase):
    id: str
    created_at: datetime


class ProductBase(BaseModel):
    name: str
    price: float
    description: str = ""


class Product(ProductBase):
    id: str
    created_at: datetime


class HealthResponse(BaseModel):
    status: str
    service: str
