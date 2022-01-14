# TODO: split E2E test to unit test
from typing import Dict, Iterator

import pytest
from sqlalchemy.orm import Session
from starlette.testclient import TestClient

import pol.server
from pol.models import Avatar, PublicUser
from pol.db.tables import ChiiRevHistory
from tests.conftest import MockUser
from pol.services.user_service import UserService
from pol.services.rev_service.person_rev import person_rev_type_filters
from pol.services.rev_service.character_rev import character_rev_type_filters

person_revisions_api_prefix = "/v0/revisions/persons"


@pytest.mark.env("e2e", "database")
def test_person_revisions_basic(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 9, person_rev_type_filters
    ):
        mock_user(r.rev_creator)
    response = client.get(person_revisions_api_prefix, params={"person_id": 9})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["data"]
    assert res["offset"] == 0
    assert "limit" in res
    for item in res["data"]:
        assert "nickname" in item["creator"]


@pytest.mark.env("e2e", "database")
def test_person_revisions_offset(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 9, person_rev_type_filters
    ):
        mock_user(r.rev_creator)
    offset = 1
    common_params = {"person_id": 9}
    response1 = client.get(
        person_revisions_api_prefix, params={"offset": 1, **common_params}
    )
    assert response1.status_code == 200
    assert response1.headers["content-type"] == "application/json"

    res = response1.json()
    assert (
        res["data"][0]["id"]
        == client.get(person_revisions_api_prefix, params=common_params).json()["data"][
            1
        ]["id"]
    )
    assert res["offset"] == offset


@pytest.mark.env("e2e", "database")
def test_person_revisions_offset_limit(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 9, person_rev_type_filters
    ):
        mock_user(r.rev_creator)
    offset = 30000
    response = client.get(
        person_revisions_api_prefix, params={"offset": offset, "person_id": 9}
    )
    assert response.status_code == 422, response.text


character_revisions_api_prefix = "/v0/revisions/characters"


@pytest.mark.env("e2e", "database")
def test_character_revisions_basic(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, character_rev_type_filters
    ):
        mock_user(r.rev_creator)
    response = client.get(character_revisions_api_prefix, params={"character_id": 1})
    assert response.status_code == 200, response.json()
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res
    assert res["data"]


@pytest.mark.env("e2e", "database")
def test_character_revisions_offset(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, character_rev_type_filters
    ):
        mock_user(r.rev_creator)
    offset = 1
    common_params = {"character_id": 1}
    response1 = client.get(
        character_revisions_api_prefix, params={"offset": offset, **common_params}
    )
    assert response1.status_code == 200
    assert response1.headers["content-type"] == "application/json"

    res = response1.json()
    assert (
        res["data"][0]["id"]
        == client.get(character_revisions_api_prefix, params=common_params).json()[
            "data"
        ][1]["id"]
    )
    assert res["offset"] == offset


@pytest.mark.env("e2e", "database")
def test_character_revisions_page_limit(
    client: TestClient,
    db_session: Session,
    mock_user: MockUser,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, character_rev_type_filters
    ):
        mock_user(r.rev_creator)
    offset = 30000
    response = client.get(
        character_revisions_api_prefix, params={"character_id": 1, "offset": offset}
    )
    assert response.status_code == 422, response.text


subject_revisions_api_prefix = "/v0/revisions/subjects"


class MockUserService:
    async def get_users_by_id(self, ids: Iterator[int]) -> Dict[int, PublicUser]:
        ids = list(ids)
        for uid in ids:
            assert uid > 0

        return {
            uid: PublicUser(
                id=uid,
                username=f"username {uid}",
                nickname=f"nickname {uid}",
                avatar=Avatar.from_db_record(""),
            )
            for uid in ids
        }

    async def get_by_uid(self, uid: int) -> PublicUser:
        assert uid > 0
        return PublicUser(
            id=uid,
            username=f"username {uid}",
            nickname=f"nickname {uid}",
            avatar=Avatar.from_db_record(""),
        )


@pytest.mark.env("e2e", "database")
def test_subject_revisions_basic(client: TestClient):
    pol.server.app.dependency_overrides[UserService.new] = MockUserService
    response = client.get(subject_revisions_api_prefix, params={"subject_id": 26})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert "total" in res
    assert "limit" in res
    assert res["offset"] == 0
    if res["total"] <= res["limit"]:
        assert res["total"] == len(res["data"])
    else:
        assert res["limit"] == len(res["data"])
    for item in res["data"]:
        if item["creator"]:
            assert "nickname" in item["creator"]


@pytest.mark.env("e2e", "database")
def test_subject_revisions_offset(client: TestClient):
    offset = 1
    common_params = {"subject_id": 1}
    response1 = client.get(
        subject_revisions_api_prefix, params={"offset": offset, **common_params}
    )
    assert response1.status_code == 200
    assert response1.headers["content-type"] == "application/json"

    res = response1.json()
    assert (
        res["data"][0]["id"]
        == client.get(subject_revisions_api_prefix, params=common_params).json()[
            "data"
        ][1]["id"]
    )
    assert res["offset"] == offset


@pytest.mark.env("e2e", "database")
def test_subject_revisions_page_limit(
    client: TestClient,
):
    offset = 30000
    response = client.get(
        subject_revisions_api_prefix, params={"subject_id": 1, "offset": offset}
    )
    assert response.status_code == 422, response.text


episode_revisions_api_prefix = "/v0/revisions/episodes"


@pytest.mark.env("e2e", "database")
def test_episode_revisions_basic(
    client: TestClient,
):
    response = client.get(episode_revisions_api_prefix, params={"episode_id": 522})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert "total" in res
    assert "limit" in res
    assert res["offset"] == 0
    if res["total"] <= res["limit"]:
        assert res["total"] == len(res["data"])
    else:
        assert res["limit"] == len(res["data"])
    for item in res["data"]:
        assert "nickname" in item["creator"]


@pytest.mark.env("e2e", "database")
def test_episode_revisions_offset(
    client: TestClient,
):
    offset = 1
    common_params = {"episode_id": 1045}
    response1 = client.get(
        episode_revisions_api_prefix, params={"offset": offset, **common_params}
    )
    assert response1.status_code == 200
    assert response1.headers["content-type"] == "application/json"

    res = response1.json()
    assert (
        res["data"][0]["id"]
        == client.get(episode_revisions_api_prefix, params=common_params).json()[
            "data"
        ][1]["id"]
    )
    assert res["offset"] == offset


@pytest.mark.env("e2e", "database")
def test_episode_revisions_page_limit(
    client: TestClient,
):
    offset = 30000
    response = client.get(
        episode_revisions_api_prefix, params={"episode_id": 522, "offset": offset}
    )
    assert response.status_code == 422, response.text
