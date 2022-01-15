import orjson
import pytest
from redis import Redis
from starlette.testclient import TestClient

from pol import config


@pytest.mark.env("e2e", "database", "redis")
def test_subject_auth_nsfw_no_auth_404(client: TestClient):
    """not authorized 404 nsfw subject"""
    response = client.get("/v0/subjects/16")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_subject_auth_nsfw(client: TestClient, auth_header):
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200
    assert response.json()["nsfw"]

    response = client.get("/v0/subjects/16")
    assert response.status_code == 404


@pytest.mark.env("e2e", "database", "redis")
def test_subject_auth_cached(client: TestClient, redis_client: Redis, auth_header):
    cache_key = config.CACHE_KEY_PREFIX + "res:subject:1"
    redis_client.set(cache_key, orjson.dumps({"id": 10}))
    response = client.get("/v0/subjects/1")
    assert response.status_code == 200, "broken cache should be purged"

    in_cache = orjson.loads(redis_client.get(cache_key))
    assert response.json()["name"] == in_cache["name"]


@pytest.mark.env("e2e", "database", "redis")
def test_subject_nsfw_no_cache(client: TestClient, auth_header):
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200
    assert response.json()["nsfw"]
    assert response.headers["cache-control"] == "no-store"
