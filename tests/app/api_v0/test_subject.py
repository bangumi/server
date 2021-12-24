import orjson.orjson
from redis import Redis
from starlette.testclient import TestClient

from pol import config


def test_subject_not_found(client: TestClient):
    response = client.get("/v0/subjects/2000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_subject_not_valid(client: TestClient):
    response = client.get("/v0/subjects/hello")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"

    response = client.get("/v0/subjects/0")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"


def test_subject_basic(client: TestClient):
    response = client.get("/v0/subjects/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["id"] == 2
    assert data["name"] == "坟场"


def test_subject_locked(client: TestClient):
    response = client.get("/v0/subjects/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["locked"]


def test_subject_nsfw_auth_200(client: TestClient, auth_header):
    """authorized 200 nsfw subject"""
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


def test_subject_redirect(client: TestClient):
    response = client.get("/v0/subjects/18", allow_redirects=False)
    assert response.status_code == 307
    assert response.headers["location"] == "/v0/subjects/19"


def test_subject_empty_image(client: TestClient, mock_subject):
    mock_subject(200)
    response = client.get("/v0/subjects/200")
    assert response.status_code == 200
    data = response.json()
    assert data["images"] is None


def test_subject_ep_query_limit_offset(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 8, "limit": 5})
    assert response.status_code == 200

    data = response.json()["data"]
    assert isinstance(data, list)
    assert len(data) == 5

    ids = [x["id"] for x in data]

    new_data = client.get(
        "/v0/episodes", params={"subject_id": 8, "limit": 4, "offset": 1}
    ).json()["data"]

    assert ids[1:] == [x["id"] for x in new_data]


def test_subject_ep_type(client: TestClient):
    response = client.get("/v0/episodes", params={"type": 3, "subject_id": 253})
    assert response.status_code == 200

    data = response.json()["data"]
    assert [x["id"] for x in data] == [103233, 103234, 103235]


def test_subject_characters(client: TestClient):
    response = client.get("/v0/subjects/8/characters")
    assert response.status_code == 200

    data = response.json()
    assert isinstance(data, list)
    assert data


def test_subject_persons(client: TestClient):
    response = client.get("/v0/subjects/4/persons")
    assert response.status_code == 200

    data = response.json()

    assert isinstance(data, list)
    assert data


def test_subject_subjects_ban(client: TestClient):
    response = client.get("/v0/subjects/5/subjects")
    assert response.status_code == 404


def test_subject_subjects(client: TestClient):
    response = client.get("/v0/subjects/11/subjects")
    assert response.status_code == 200
    data = response.json()

    assert isinstance(data, list)
    assert data


def test_subject_cache_broken_purge(client: TestClient, redis_client: Redis):
    cache_key = config.CACHE_KEY_PREFIX + "subject:1"
    redis_client.set(cache_key, orjson.dumps({"id": 10, "test": "1"}))
    response = client.get("/v0/subjects/1")
    assert response.status_code == 200, "broken cache should be purged"

    in_cache = orjson.loads(redis_client.get(cache_key))
    assert response.json()["name"] == in_cache["name"]
    assert "test" not in in_cache


def test_subject_tags(client: TestClient):
    response = client.get("/v0/subjects/2")
    assert response.json()["tags"] == [
        {"name": "陈绮贞", "count": 9},
        {"name": "中配", "count": 1},
        {"name": "银魂中配", "count": 1},
        {"name": "神还原", "count": 1},
        {"name": "冷泉夜月", "count": 1},
        {"name": "银他妈", "count": 1},
        {"name": "陈老师", "count": 1},
        {"name": "银魂", "count": 1},
        {"name": "治愈系", "count": 1},
        {"name": "恶搞", "count": 1},
    ]


def test_subject_tags_empty(client: TestClient, mock_subject):
    sid = 15234523
    mock_subject(sid)
    response = client.get(f"/v0/subjects/{sid}")
    assert response.json()["tags"] == []
