from datetime import datetime, timedelta

import orjson
import pytest
from redis import Redis
from fastapi import FastAPI
from starlette.testclient import TestClient

from pol import config, models
from tests.base import async_lambda
from pol.db.const import SubjectType
from pol.db.tables import ChiiSubject
from pol.permission import UserGroup
from pol.models.user import User
from tests.fixtures.mock_db import MockAsyncSession
from pol.api.v0.depends.auth import Guest, optional_user
from tests.fixtures.mock_service import MockSubjectService


def test_subject_auth_nsfw_no_auth_404(
    client: TestClient,
    auth_header,
    mock_redis,
    mock_db: MockAsyncSession,
    app: FastAPI,
    mock_subject_service: MockSubjectService,
):
    """not authorized 404 nsfw subject"""
    mock_subject_service.get_by_id.return_value = models.Subject(
        id=16, type=SubjectType.anime, nsfw=True, platform=1, redirect=0, ban=0
    )
    mock_db.scalar.return_value = 10  # episode count
    app.dependency_overrides[optional_user] = async_lambda(Guest())

    response = client.get("/v0/subjects/16")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_subject_auth_nsfw(
    client: TestClient,
    auth_header,
    mock_redis,
    mock_db: MockAsyncSession,
    app: FastAPI,
    mock_subject_service: MockSubjectService,
):
    mock_subject_service.get_by_id.return_value = models.Subject(
        id=16, type=SubjectType.anime, nsfw=True, platform=1, redirect=0, ban=0
    )
    mock_db.scalar.return_value = 10  # episode count
    mock_db.get.return_value = ChiiSubject.with_default_value(subject_nsfw=1)

    app.dependency_overrides[optional_user] = async_lambda(
        User(
            id=1,
            username="",
            nickname="",
            group_id=UserGroup.normal_user,
            registration_date=datetime.now().astimezone() - timedelta(days=90),
        )
    )
    response = client.get("/v0/subjects/16")
    assert response.status_code == 200
    assert response.json()["nsfw"]


@pytest.mark.env("e2e", "database", "redis")
def test_subject_auth_cached(client: TestClient, redis_client: Redis, auth_header):
    cache_key = config.CACHE_KEY_PREFIX + "res:subject:1"
    redis_client.set(cache_key, orjson.dumps({"id": 10}))
    response = client.get("/v0/subjects/1")
    assert response.status_code == 200, "broken cache should be purged"

    in_cache = orjson.loads(redis_client.get(cache_key))
    assert response.json()["name"] == in_cache["name"]


def test_subject_nsfw_no_cache(
    client: TestClient,
    auth_header,
    mock_redis,
    mock_db: MockAsyncSession,
    app: FastAPI,
    mock_subject_service: MockSubjectService,
):
    app.dependency_overrides[optional_user] = async_lambda(
        User(
            id=1,
            username="",
            nickname="",
            group_id=UserGroup.normal_user,
            registration_date=datetime.now().astimezone() - timedelta(days=90),
        )
    )
    mock_subject_service.get_by_id.return_value = models.Subject(
        id=16, type=SubjectType.anime, nsfw=True, platform=1, redirect=0, ban=0
    )
    mock_db.scalar.return_value = 10  # episode count
    mock_db.get.return_value = ChiiSubject.with_default_value(subject_nsfw=1)
    response = client.get("/v0/subjects/16", headers=auth_header)
    assert response.status_code == 200
    assert response.json()["nsfw"]
    # assert response.headers["cache-control"] == "no-store"
