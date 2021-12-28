import orjson.orjson
from redis import Redis
from starlette.testclient import TestClient

from pol import config


def test_subject_auth_nsfw_no_auth_404(client: TestClient):
    """not authorized 404 nsfw subject"""
    response = client.get("/v0/subjects/16")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_subject_auth_nsfw(client: TestClient, auth_header):
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200

    response = client.get("/v0/subjects/16")
    assert response.status_code == 404


def test_subject_auth_cached(client: TestClient, redis_client: Redis, auth_header):
    cache_key = config.CACHE_KEY_PREFIX + "subject:1"
    redis_client.set(cache_key, orjson.dumps({"id": 10}))
    response = client.get("/v0/subjects/1")
    assert response.status_code == 200, "broken cache should be pruged"

    in_cache = orjson.loads(redis_client.get(cache_key))
    assert response.json()["name"] == in_cache["name"]


def test_subject_nsfw_no_cache(client: TestClient, auth_header):
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["cache-control"] == "no-store"
