from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.db.tables import ChiiRevHistory
from tests.conftest import MockAccessToken
from pol.api.v0.revision import person_rev_type_filters, character_rev_type_filters

person_revisions_api_prefix = "/v0/revisions/persons"


def test_person_revisions_basic(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 9, person_rev_type_filters
    ):
        mock_access_token(r.rev_creator)
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


def test_person_revisions_offset(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 9, person_rev_type_filters
    ):
        mock_access_token(r.rev_creator)
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


def test_person_revisions_offset_limit(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 9, person_rev_type_filters
    ):
        mock_access_token(r.rev_creator)
    offset = 30000
    response = client.get(
        person_revisions_api_prefix, params={"offset": offset, "person_id": 9}
    )
    assert response.status_code == 422, response.text


character_revisions_api_prefix = "/v0/revisions/characters"


def test_character_revisions_basic(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, character_rev_type_filters
    ):
        mock_access_token(r.rev_creator)
    response = client.get(character_revisions_api_prefix, params={"charater_id": 1})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res
    for item in res["data"]:
        assert "nickname" in item["creator"]


def test_character_revisions_offset(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, character_rev_type_filters
    ):
        mock_access_token(r.rev_creator)
    offset = 1
    common_params = {"charater_id": 1}
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


def test_character_revisions_page_limit(
    client: TestClient,
    db_session: Session,
    mock_access_token: MockAccessToken,
):
    for r in db_session.query(ChiiRevHistory.rev_creator).where(
        ChiiRevHistory.rev_mid == 1, character_rev_type_filters
    ):
        mock_access_token(r.rev_creator)
    offset = 30000
    response = client.get(
        character_revisions_api_prefix, params={"charater_id": 1, "offset": offset}
    )
    assert response.status_code == 422, response.text
