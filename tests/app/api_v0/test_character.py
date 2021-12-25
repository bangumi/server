import json

import redis
from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.config import CACHE_KEY_PREFIX
from pol.db.tables import ChiiRevHistory
from tests.conftest import MockAccessToken
from pol.api.v0.character import rev_type_filters


def test_character_not_found(client: TestClient):
    response = client.get("/v0/characters/2000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_character_not_valid(client: TestClient):
    response = client.get("/v0/characters/hello")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"

    response = client.get("/v0/characters/0")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"


def test_character_basic(client: TestClient):
    response = client.get("/v0/characters/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert not data["locked"]


def test_character_locked(client: TestClient):
    response = client.get("/v0/characters/9")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["locked"]


def test_character_ban_404(client: TestClient):
    response = client.get("/v0/characters/6")
    assert response.status_code == 404


def test_character_cache(client: TestClient, redis_client: redis.Redis):
    response = client.get("/v0/characters/1")
    assert response.status_code == 200
    assert response.headers["x-cache-status"] == "miss"

    response = client.get("/v0/characters/1")
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200

    cache_key = CACHE_KEY_PREFIX + "character:1"

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
    response = client.get("/v0/characters/1")
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200

    res = response.json()
    assert res["name"] == "n"


def test_character_subjects(client: TestClient):
    response = client.get("/v0/characters/1/subjects")
    assert response.status_code == 200

    subjects = response.json()
    assert subjects[0]["id"] == 8
    assert set(subjects[0].keys()) == {
        "id",
        "staff",
        "name",
        "name_cn",
        "image",
    }


def test_character_subjects_ban(client: TestClient):
    response = client.get("/v0/characters/55/subjects", allow_redirects=False)
    assert response.status_code == 404


def test_character_redirect(client: TestClient):
    response = client.get("/v0/characters/55", allow_redirects=False)
    assert response.status_code == 307
    assert response.headers["location"] == "/v0/characters/52"


def test_character_lock(client: TestClient):
    response = client.get("/v0/characters/9")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["locked"]


def test_character_persons(client: TestClient, mock_person):
    mock_person(3818, "福山潤")
    response = client.get("/v0/characters/1/persons")
    assert response.status_code == 200

    persons = response.json()
    assert persons[0]["id"] == 3818
    assert persons[0]["subject_id"] == 8
    assert set(persons[0].keys()) == {
        "id",
        "name",
        "type",
        "images",
        "subject_id",
        "subject_name",
        "subject_name_cn",
    }


def test_character_revisions_basic(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, rev_type_filters
    ):
        mock_access_token(r["rev_creator"])
    response = client.get(
        "/v0/characters/1/revisions",
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res
    for item in res["data"]:
        assert "nickname" in item["creator"]


def test_character_revisions_filter_uid(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, rev_type_filters
    ):
        mock_access_token(r["rev_creator"])
    uid = 1
    response = client.get("/v0/characters/1/revisions", params={"uid": uid})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["data"]
    assert res["offset"] == 0
    assert "limit" in res
    for item in res["data"]:
        assert item["creator"]["id"] == uid
        assert "nickname" not in item["creator"]


def test_character_revisions_offset(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, rev_type_filters
    ):
        mock_access_token(r["rev_creator"])
    offset = 1
    response1 = client.get("/v0/characters/1/revisions", params={"offset": 1})
    assert response1.status_code == 200
    assert response1.headers["content-type"] == "application/json"

    res = response1.json()
    assert (
        res["data"][0]["id"]
        == client.get("/v0/characters/1/revisions").json()["data"][1]["id"]
    )
    assert res["offset"] == offset


def test_character_revisions_page_limit(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, rev_type_filters
    ):
        mock_access_token(r["rev_creator"])
    offset = 30000
    response = client.get("/v0/characters/1/revisions", params={"offset": offset})
    assert response.status_code == 422, response.text


def test_character_revision_basic(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_id == 190704
    ):
        mock_access_token(r["rev_creator"])
    response = client.get(
        "/v0/characters/1/revisions/190704",
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res: dict = response.json()
    assert set(res.keys()) == {
        "id",
        "type",
        "timestamp",
        "summary",
        "data",
        "creator",
    }
    assert res["data"]["1"]["crt_name"] == "ルルーシュ・ランペルージ"


def test_character_revision_cache(
    client: TestClient,
    redis_client: redis.Redis,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_id == 190704
    ):
        mock_access_token(r["rev_creator"])
    response = client.get(
        "/v0/characters/1/revisions/190704",
    )
    assert response.status_code == 200
    assert response.headers["x-cache-status"] == "miss"

    response = client.get(
        "/v0/characters/1/revisions/190704",
    )
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200
