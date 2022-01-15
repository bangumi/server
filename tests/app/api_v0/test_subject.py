from pathlib import Path
from unittest import mock
from unittest.mock import AsyncMock

import orjson
import pytest
from redis import Redis
from fastapi import FastAPI
from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol import config
from pol.db import sa
from pol.models import Subject
from tests.base import async_lambda
from pol.db.tables import ChiiSubjectField
from tests.conftest import MockRedis
from pol.api.v0.depends import optional_user
from pol.api.v0.depends.auth import Guest
from pol.services.subject_service import SubjectService

fixtures_path = Path(__file__).parent.joinpath("fixtures")


def test_subject_not_found(
    client: TestClient, app: FastAPI, mock_redis: MockRedis, mock_db
):
    app.dependency_overrides[SubjectService.new] = async_lambda(
        mock.Mock(get_by_id=AsyncMock(side_effect=SubjectService.NotFoundError))
    )
    app.dependency_overrides[optional_user] = async_lambda(Guest())

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


@pytest.mark.env("e2e", "database", "redis")
def test_subject_basic(client: TestClient):
    response = client.get("/v0/subjects/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["id"] == 2
    assert data["name"] == "坟场"
    assert not response.json()["nsfw"]
    assert data["locked"]


@pytest.mark.env("e2e", "database", "redis")
def test_subject_nsfw_auth_200(client: TestClient, auth_header):
    """authorized 200 nsfw subject"""
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


def test_subject_redirect(
    client: TestClient, app: FastAPI, mock_redis: MockRedis, mock_db
):
    app.dependency_overrides[SubjectService.new] = async_lambda(
        mock.Mock(
            get_by_id=AsyncMock(
                return_value=Subject(
                    id=18,
                    type=2,
                    name="",
                    name_cn="",
                    summary="",
                    nsfw=False,
                    platform=1,
                    image="",
                    infobox="",
                    redirect=19,
                    ban=0,
                )
            )
        )
    )
    app.dependency_overrides[optional_user] = async_lambda(Guest())

    response = client.get("/v0/subjects/18", allow_redirects=False)
    assert response.status_code == 307
    assert response.headers["location"] == "/v0/subjects/19"
    assert response.headers["cache-control"] == "public, max-age=300"


@pytest.mark.env("e2e", "database", "redis")
def test_subject_empty_image(client: TestClient, mock_subject):
    mock_subject(200)
    response = client.get("/v0/subjects/200")
    assert response.status_code == 200
    data = response.json()
    assert data["images"] is None


@pytest.mark.env("e2e", "database", "redis")
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


@pytest.mark.env("e2e", "database", "redis")
def test_subject_ep_type(client: TestClient):
    response = client.get("/v0/episodes", params={"type": 3, "subject_id": 253})
    assert response.status_code == 200

    data = response.json()["data"]
    assert [x["id"] for x in data] == [103233, 103234, 103235]


@pytest.mark.env("e2e", "database", "redis")
def test_subject_characters(client: TestClient):
    response = client.get("/v0/subjects/8/characters")
    assert response.status_code == 200

    data = response.json()
    assert isinstance(data, list)
    assert data


@pytest.mark.env("e2e", "database", "redis")
def test_subject_persons(client: TestClient):
    response = client.get("/v0/subjects/4/persons")
    assert response.status_code == 200

    data = response.json()

    assert isinstance(data, list)
    assert data


@pytest.mark.env("e2e", "database", "redis")
def test_subject_subjects_ban(client: TestClient):
    response = client.get("/v0/subjects/5/subjects")
    assert response.status_code == 404


@pytest.mark.env("e2e", "database", "redis")
def test_subject_subjects(client: TestClient):
    response = client.get("/v0/subjects/11/subjects")
    assert response.status_code == 200
    data = response.json()

    assert isinstance(data, list)
    assert data


@pytest.mark.env("e2e", "database", "redis")
def test_subject_cache_broken_purge(client: TestClient, redis_client: Redis):
    cache_key = config.CACHE_KEY_PREFIX + "res:subject:1"
    redis_client.set(cache_key, orjson.dumps({"id": 10, "test": "1"}))
    response = client.get("/v0/subjects/1")
    assert response.status_code == 200, "broken cache should be purged"

    in_cache = orjson.loads(redis_client.get(cache_key))
    assert response.json()["name"] == in_cache["name"]
    assert "test" not in in_cache


@pytest.mark.env("e2e", "database", "redis")
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


@pytest.mark.env("e2e", "database", "redis")
def test_subject_tags_empty(client: TestClient, mock_subject):
    sid = 15234523
    mock_subject(sid)
    response = client.get(f"/v0/subjects/{sid}")
    assert response.json()["tags"] == []


@pytest.mark.env("e2e", "database", "redis")
def test_subject_tags_none(client: TestClient, mock_subject, db_session: Session):
    """
    should exclude a tag if name is None.
    todo: can count be None too?
    """
    sid = 15234524
    mock_subject(sid)
    field_tags_with_none_val = (
        fixtures_path.joinpath("subject_2585_tags.txt").read_bytes().strip()
    )
    db_session.execute(
        sa.update(ChiiSubjectField)
        .where(ChiiSubjectField.field_sid == sid)
        .values(field_tags=field_tags_with_none_val)
    )
    db_session.commit()
    response = client.get(f"/v0/subjects/{sid}")
    assert response.json()["tags"] == [
        {"name": "炮姐", "count": 1956},
        {"name": "超电磁炮", "count": 1756},
        {"name": "J.C.STAFF", "count": 1746},
        {"name": "御坂美琴", "count": 1367},
        {"name": "百合", "count": 1240},
        {"name": "2009年10月", "count": 917},
        {"name": "bilibili", "count": 795},
        {"name": "TV", "count": 709},
        {"name": "黑子", "count": 702},
        {"name": "科学超电磁炮", "count": 621},
        {"name": "魔法禁书目录", "count": 518},
        {"name": "2009", "count": 409},
        {"name": "漫画改", "count": 288},
        {"name": "傲娇娘", "count": 280},
        {"name": "校园", "count": 156},
        {"name": "战斗", "count": 144},
        {"name": "长井龙雪", "count": 123},
        {"name": "漫改", "count": 110},
        {"name": "姐控", "count": 107},
        {"name": "轻小说改", "count": 93},
        {"name": "科幻", "count": 82},
        {"name": "超能力", "count": 73},
        {"name": "日常", "count": 58},
        {"name": "奇幻", "count": 54},
        {"name": "豊崎愛生", "count": 53},
        {"name": "長井龍雪", "count": 47},
        {"name": "某科学的超电磁炮", "count": 47},
        {"name": "佐藤利奈", "count": 38},
        {"name": "新井里美", "count": 34},
    ]


@pytest.mark.env("e2e", "database", "redis")
def test_subject_cache_header_public(client: TestClient, redis_client: Redis):
    response = client.get("/v0/subjects/1")
    assert response.status_code == 200, "broken cache should be purged"

    assert response.headers["cache-control"] == "public, max-age=300"
    assert not response.json()["nsfw"]
