from pathlib import Path

# import orjson.orjson
# from redis import Redis
# from sqlalchemy.orm import Session
import pytest
from starlette.testclient import TestClient

# from pol import sa, config
# from pol.db.tables import ChiiSubjectField

fixtures_path = Path(__file__).parent.joinpath("fixtures")


subject_ids = {
    "public": 1,
    "nsfw": 16,
    "banned": 5,
    "locked": 2
}


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_guest_normal(client: TestClient):
    """returns a list of topics that belongs to a topic"""
    subject_id = subject_ids["public"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 200

    data = response.json()
    assert isinstance(data["data"], list)
    assert len(data["data"]) == 1


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_guest_nsfw(client: TestClient):
    """subject is nsfw, user not logged in"""
    subject_id = subject_ids["nsfw"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_user_age_insuf_nsfw(client: TestClient):
    """subject is nsfw, user reg days <= 60"""
    # todo: add user
    subject_id = subject_ids["nsfw"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_user_age_pass_nsfw(client: TestClient):
    """subject is nsfw, user reg days > 60"""
    # todo: add user
    subject_id = subject_ids["nsfw"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_guest_locked(client: TestClient):
    """subject is locked, user not logged in"""
    subject_id = subject_ids["locked"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_user_locked(client: TestClient):
    """subject is locked, user logged in"""
    # todo: add user
    subject_id = subject_ids["locked"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_guest_ban(client: TestClient):
    """subject is banned, user not logged in"""
    subject_id = subject_ids["banned"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_user_age_insuf_ban(client: TestClient):
    """subject is locked, user logged in"""
    # todo: add user
    subject_id = subject_ids["banned"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_user_age_pass_locked(client: TestClient):
    """subject is nsfw, user logged in, reg days > 10 yr"""
    # todo: add user
    subject_id = subject_ids["locked"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404


@pytest.mark.skip(reason="wip")
def test_subject_topics_authz_user_age_pass_ban(client: TestClient):
    """subject is nsfw, user logged in, reg days > 10 yr"""
    # todo: add user
    subject_id = subject_ids["banned"]
    response = client.get(f"/v0/subjects/{subject_id}/topics")
    assert response.status_code == 404
