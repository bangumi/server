# TODO: split E2E test to unit test
import pytest
from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.db.tables import ChiiRevHistory
from tests.conftest import MockUser

person_revisions_api_prefix = "/v0/revisions/persons"


@pytest.mark.env("e2e")
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


@pytest.mark.env("e2e")
def test_person_revision_not_found(client: TestClient):
    response = client.get(
        f"{person_revisions_api_prefix}/888888",
    )
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


character_revisions_api_prefix = "/v0/revisions/characters"


@pytest.mark.env("e2e")
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


@pytest.mark.env("e2e")
def test_character_revision_not_found(client: TestClient):
    response = client.get(
        f"{character_revisions_api_prefix}/888888",
    )
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


subject_revisions_api_prefix = "/v0/revisions/subjects"


@pytest.mark.env("e2e")
def test_subject_revision_basic(
    client: TestClient,
):
    response = client.get(
        f"{subject_revisions_api_prefix}/718391",
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
    assert res["data"]["name"] == "第一次的親密接觸"


@pytest.mark.env("e2e")
def test_subject_revision_not_found(client: TestClient):
    response = client.get(
        f"{subject_revisions_api_prefix}/888888",
    )
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e")
def test_subject_revision_amazon(client: TestClient):
    response = client.get(
        f"{subject_revisions_api_prefix}/194043",
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
    assert res["creator"] is None


episode_revisions_api_prefix = "/v0/revisions/episodes"


@pytest.mark.env("e2e")
def test_episode_revision_basic(
    client: TestClient,
):
    response = client.get(
        f"{episode_revisions_api_prefix}/1435",
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
    assert (
        res["data"]["eids"]
        == "522,523,524,525,526,527,528,529,530,531,532,89,90,2,91,104,374,520,574,577"
    )


@pytest.mark.env("e2e")
def test_episode_revision_not_found(client: TestClient):
    response = client.get(
        f"{episode_revisions_api_prefix}/888888",
    )
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"
