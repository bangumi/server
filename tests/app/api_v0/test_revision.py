import redis
from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.db.tables import ChiiRevHistory
from tests.conftest import MockUser

person_revisions_api_prefix = "/v0/revisions/persons"


def test_person_revision_basic(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_id == 348475
    ):
        mock_user(r.rev_creator)
    response = client.get(
        f"{person_revisions_api_prefix}/348475",
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res: dict = response.json()
    assert set(res.keys()) == {
        "id",
        "type",
        "created_at",
        "summary",
        "data",
        "creator",
    }
    assert res["data"]["348475"]["prsn_name"] == "森岡浩之"


def test_person_revision_cache(
    client: TestClient,
    redis_client: redis.Redis,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_id == 348475
    ):
        mock_user(r.rev_creator)
    url = f"{person_revisions_api_prefix}/348475"
    response = client.get(
        url,
    )
    assert response.status_code == 200
    assert response.headers["x-cache-status"] == "miss"

    response = client.get(
        url,
    )
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200


character_revisions_api_prefix = "/v0/revisions/characters"


def test_character_revision_basic(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_id == 190704
    ):
        mock_user(r.rev_creator)
    response = client.get(
        f"{character_revisions_api_prefix}/190704",
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res: dict = response.json()
    assert set(res.keys()) == {
        "id",
        "type",
        "created_at",
        "summary",
        "data",
        "creator",
    }
    assert res["data"]["1"]["crt_name"] == "ルルーシュ・ランペルージ"


def test_character_revision_cache(
    client: TestClient,
    redis_client: redis.Redis,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_id == 190704
    ):
        mock_user(r.rev_creator)
    url = f"{character_revisions_api_prefix}/190704"
    response = client.get(
        url,
    )
    assert response.status_code == 200
    assert response.headers["x-cache-status"] == "miss"

    response = client.get(
        url,
    )
    assert response.headers["x-cache-status"] == "hit"
    assert response.status_code == 200
