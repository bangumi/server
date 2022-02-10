from typing import Any, Dict, Iterator, cast

import pytest
from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.db import sa
from pol.db.tables import ChiiIndexRelated
from tests.fixtures.mock_service import MockUserService
from tests.fixtures.mock_db_record import MockSubject

index_api_prefix = "/v0/indices"


def mock_subjects(
    db_session: Session,
    mock_subject: MockSubject,
    id: int,
    type: int = 0,
    offset=0,
    limit=30,
    is_nsfw=True,
):
    where = [ChiiIndexRelated.idx_rlt_rid == id]
    if type != 0:
        where.append(ChiiIndexRelated.idx_rlt_type == type)
    for r in cast(
        Iterator[ChiiIndexRelated],
        db_session.scalars(
            sa.select(ChiiIndexRelated).where(*where).offset(offset).limit(limit)
        ),
    ):
        try:
            mock_subject(r.idx_rlt_sid, subject_nsfw=int(is_nsfw))
        except ValueError:
            pass


@pytest.mark.env("e2e", "database")
def test_index(client: TestClient, mock_user_service: MockUserService):
    response = client.get(
        f"{index_api_prefix}/15045",
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res: Dict[str, Any] = response.json()
    assert set(res.keys()) == {
        "id",
        "title",
        "desc",
        "stat",
        "total",
        "created_at",
        "creator",
        "ban",
    }


@pytest.mark.env("e2e", "database", "redis")
def test_index_nsfw_404(
    client: TestClient,
    mock_user_service: MockUserService,
    db_session: Session,
    mock_subject: MockSubject,
):
    id = 15465
    mock_subjects(db_session, mock_subject, id)
    response = client.get(
        f"{index_api_prefix}/{id}",
    )
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_index_nsfw_200(
    client: TestClient,
    mock_user_service: MockUserService,
    auth_header: Dict[str, str],
    db_session: Session,
    mock_subject: MockSubject,
):
    id = 15465
    mock_subjects(db_session, mock_subject, id)
    response = client.get(f"{index_api_prefix}/{id}", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database")
def test_index_subjects(
    client: TestClient,
    db_session: Session,
    mock_subject: MockSubject,
):
    id = 15045
    mock_subjects(db_session, mock_subject, id, is_nsfw=False)
    response = client.get(
        f"{index_api_prefix}/{id}/subjects",
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["data"]
    assert res["offset"] == 0
    assert "limit" in res

    for item in res["data"]:
        assert "comment" in item


@pytest.mark.env("e2e", "database", "redis")
def test_index_subjects_nsfw_404(
    client: TestClient,
    mock_user_service: MockUserService,
    db_session: Session,
    mock_subject: MockSubject,
):
    id = 15465
    mock_subjects(db_session, mock_subject, id)

    response = client.get(
        f"{index_api_prefix}/{id}/subjects",
    )
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_index_subjects_nsfw_200(
    client: TestClient,
    mock_user_service: MockUserService,
    auth_header: Dict[str, str],
    db_session: Session,
    mock_subject: MockSubject,
):
    id = 15465
    mock_subjects(db_session, mock_subject, id)
    response = client.get(f"{index_api_prefix}/{id}/subjects", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_index_subjects_type(
    client: TestClient,
    db_session: Session,
    mock_subject: MockSubject,
):
    id = 15045
    type = 2
    mock_subjects(db_session, mock_subject, id, type=type, is_nsfw=False)
    response = client.get(f"{index_api_prefix}/{id}/subjects?type={type}")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["data"]
    for item in res["data"]:
        assert item["type"] == 2

    type = 3
    mock_subjects(db_session, mock_subject, id, type=type, is_nsfw=False)
    response = client.get(f"{index_api_prefix}/{id}/subjects?type={type}")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"
