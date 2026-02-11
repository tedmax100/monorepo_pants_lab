"""User Service — FastAPI application for managing users."""

from __future__ import annotations

from datetime import datetime
from uuid import uuid4

from fastapi import FastAPI, HTTPException

from common.models import HealthResponse, User, UserBase

app = FastAPI(title="User Service", version="0.1.0")

# In-memory store
_users: dict[str, User] = {}


@app.get("/health", response_model=HealthResponse)
def health():
    return HealthResponse(status="ok", service="user-service")


@app.get("/users", response_model=list[User])
def list_users():
    return list(_users.values())


@app.get("/users/{user_id}", response_model=User)
def get_user(user_id: str):
    if user_id not in _users:
        raise HTTPException(status_code=404, detail="User not found")
    return _users[user_id]


@app.post("/users", response_model=User, status_code=201)
def create_user(payload: UserBase):
    user = User(
        id=str(uuid4()),
        name=payload.name,
        email=payload.email,
        created_at=datetime.now(),
    )
    _users[user.id] = user
    return user
