import datetime

import orjson
import pytest
from aioredis import Redis
from starlette.testclient import TestClient

from pol import config
from pol.models import User, Avatar, UserGroup, Permission

access_token = "a_development_access_token"


@pytest.mark.env("e2e", "database", "redis")
def test_auth_200(client: TestClient, auth_header):
    response = client.get("/v0/me", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_auth_403(client: TestClient):
    response = client.get("/v0/me", headers={"Authorization": "Bearer "})
    assert response.status_code == 403, "no token"

    response = client.get("/v0/me", headers={"Authorization": f"t {access_token}"})
    assert response.status_code == 403, "no token"


@pytest.mark.env("e2e", "database", "redis")
def test_auth_403_wrong_token(client: TestClient):
    response = client.get("/v0/me", headers={"Authorization": "Bearer 1"})
    assert response.status_code == 403, "no token"


@pytest.mark.env("e2e", "database", "redis")
def test_auth_cached(client: TestClient, redis_client: Redis):
    cache_key = config.CACHE_KEY_PREFIX + "access:1"
    u = User(
        id=10,
        username="ua",
        registration_date=datetime.datetime(2007, 8, 10, 3, 1, 5),
        group_id=UserGroup.wiki_admin,
        nickname="ni",
        sign="",
        avatar=Avatar.from_db_record(""),
        permission=Permission(),
    )
    redis_client.set(cache_key, u.json(by_alias=True))
    response = client.get("/v0/me", headers={"Authorization": "Bearer 1"})
    assert response.status_code == 200, "user lookup should be cached"
    assert (
        response.json()["avatar"]["large"] == "https://lain.bgm.tv/pic/user/l/icon.jpg"
    )


@pytest.mark.env("e2e", "database", "redis")
def test_auth_cache_ban_cache_fallback(client: TestClient, redis_client: Redis):
    cache_key = config.CACHE_KEY_PREFIX + f"access:{access_token}"
    redis_client.set(
        cache_key,
        orjson.dumps(
            {
                "id": 10,
                "username": 1,
                "regdate": "2007-08-10T03:01:05",
            }
        ),
    )
    response = client.get("/v0/me", headers={"Authorization": f"Bearer {access_token}"})
    assert response.status_code == 200, "error cache should callback to db lookup"
