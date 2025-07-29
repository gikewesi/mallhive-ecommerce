# tests.py

import pytest
from httpx import AsyncClient
from main import app


@pytest.mark.asyncio
async def test_register_user():
    user_data = {
        "username": "testuser",
        "first_name": "Test",
        "last_name": "User",
        "email": "testuser@example.com",
        "phone_number": "1234567890",
        "gender": "other",
        "address": "123 Test Lane",
        "password": "StrongPassword123!"
    }

    async with AsyncClient(app=app, base_url="http://testserver") as client:
        response = await client.post("/register", json=user_data)

    assert response.status_code == 200
    assert "message" in response.json()


@pytest.mark.asyncio
async def test_login_user():
    async with AsyncClient(app=app, base_url="http://testserver") as client:
        response = await client.post("/login", data={
            "username": "testuser@example.com",
            "password": "StrongPassword123!"
        })

    assert response.status_code == 200
    assert "access_token" in response.json()


@pytest.mark.asyncio
async def test_verify_email_invalid_code():
    async with AsyncClient(app=app, base_url="http://testserver") as client:
        response = await client.post("/verify-email", json={
            "email": "testuser@example.com",
            "code": "invalid"
        })

    assert response.status_code == 400
    assert response.json()["detail"] == "Invalid verification code"
