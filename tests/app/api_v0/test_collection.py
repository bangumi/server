import pytest
from starlette.testclient import TestClient

from pol.db.const import SubjectType, CollectionType
from tests.conftest import MockUser


@pytest.mark.env("e2e", "database", "redis")
def test_collection_not_found(client: TestClient):
    response = client.get("/v0/users/2000000/collections")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


@pytest.mark.env("e2e", "database", "redis")
def test_collection_public(client: TestClient, mock_user_collection):
    mock_user_collection(id=1, user_id=382951, subject_id=1, private=True)
    response = client.get("/v0/users/382951/collections")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert len(response.json()["data"]) == 5


@pytest.mark.env("e2e", "database", "redis")
def test_collection_private(client: TestClient, auth_header, mock_user_collection):
    mock_user_collection(id=1, user_id=382951, subject_id=1, private=False)
    response = client.get("/v0/users/382951/collections", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert len(response.json()["data"]) == 6


@pytest.mark.env("e2e", "database", "redis")
def test_collection_username(
    client: TestClient,
    auth_header,
    mock_user_collection,
    mock_user: MockUser,
):
    mock_user(user_id=6, username="ua")
    mock_user_collection(id=1, user_id=6, subject_id=1, private=False)
    response = client.get("/v0/users/ua/collections", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert len(response.json()["data"]) == 1


@pytest.mark.env("e2e", "database", "redis")
def test_collection_filter_subject(
    client: TestClient, auth_header, mock_user_collection
):
    mock_user_collection(
        id=1,
        user_id=382951,
        subject_id=100,
        subject_type=SubjectType.anime,
        private=False,
    )
    response = client.get(
        "/v0/users/382951/collections", params={"subject_type": 2}, headers=auth_header
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert 100 in [x["subject_id"] for x in response.json()["data"]]
    for x in response.json()["data"]:
        assert x["subject_type"] == 2
    assert not [x for x in response.json()["data"] if x["subject_type"] != 2]


@pytest.mark.env("e2e", "database", "redis")
def test_collection_filter_type(client: TestClient, auth_header, mock_user_collection):
    mock_user_collection(
        id=1,
        user_id=382951,
        subject_id=100,
        subject_type=SubjectType.anime,
        type=CollectionType.doing,
        private=False,
    )
    response = client.get(
        "/v0/users/382951/collections",
        params={"type": CollectionType.doing.value},
        headers=auth_header,
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert 100 in [x["subject_id"] for x in response.json()["data"]]
    for x in response.json()["data"]:
        assert x["type"] == CollectionType.doing.value
    assert not [
        x for x in response.json()["data"] if x["type"] != CollectionType.doing.value
    ]
