import json

import redis
import pytest
from starlette.testclient import TestClient

from pol.config import CACHE_KEY_PREFIX


@pytest.mark.env("e2e", "database", "redis")
def test_person_not_found(client: TestClient):
    response = client.get("/v0/persons/2000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_person_not_valid(client: TestClient):
    response = client.get("/v0/persons/hello")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"

    response = client.get("/v0/persons/0")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_person_basic(client: TestClient):
    response = client.get("/v0/persons/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["img"] is None
    assert not data["locked"]


@pytest.mark.env("e2e", "database", "redis")
def test_person_ban_404(client: TestClient):
    response = client.get("/v0/persons/6")
    assert response.status_code == 404


@pytest.mark.env("e2e", "database", "redis")
def test_person_cache(client: TestClient, redis_client: redis.Redis):
    response = client.get("/v0/persons/1")
    assert response.status_code == 200
    assert response.headers["x-cache-status"] == "miss"

    response = client.get("/v0/persons/1")
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200

    cache_key = CACHE_KEY_PREFIX + "person:1"

    cached_data = {
        "id": 1,
        "name": "n",
        "type": 1,
        "career": [],
        "locked": False,
        "last_modified": 10,
        "summary": "s",
        "stat": {"comments": 110, "collects": 841},
    }
    redis_client.set(cache_key, json.dumps(cached_data))
    response = client.get("/v0/persons/1")
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200

    res = response.json()
    assert res["name"] == "n"


@pytest.mark.env("e2e", "database", "redis")
def test_person_subjects(client: TestClient):
    response = client.get("/v0/persons/1/subjects")
    assert response.status_code == 200

    subjects = response.json()
    assert set(subjects[0].keys()) == {
        "id",
        "staff",
        "name",
        "name_cn",
        "image",
    }


@pytest.mark.env("e2e", "database", "redis")
def test_person_redirect(client: TestClient):
    response = client.get("/v0/persons/10", allow_redirects=False)
    assert response.status_code == 307


@pytest.mark.env("e2e", "database", "redis")
def test_person_lock(client: TestClient):
    response = client.get("/v0/persons/9")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["locked"]


@pytest.mark.env("e2e", "database", "redis")
def test_person_characters(client: TestClient, mock_person):
    mock_person(3818, "福山潤")
    response = client.get("/v0/persons/3818/characters")
    assert response.status_code == 200

    characters = response.json()
    assert characters[0]["id"] == 1
    assert characters[0]["subject_id"] == 8
    assert set(characters[0].keys()) == {
        "id",
        "name",
        "type",
        "images",
        "subject_id",
        "subject_name",
        "subject_name_cn",
    }
