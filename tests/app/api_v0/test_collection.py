from starlette.testclient import TestClient


def test_collection_not_found(client: TestClient):
    response = client.get("/v0/user/2000000/collections")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_collection_public(client: TestClient, mock_user_collection):
    mock_user_collection(id=1, uid=382951, subject_id=1, private=True)
    response = client.get("/v0/user/382951/collections")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert len(response.json()["data"]) == 5


def test_collection_private(client: TestClient, auth_header, mock_user_collection):
    mock_user_collection(id=1, uid=382951, subject_id=1, private=False)
    response = client.get("/v0/user/382951/collections", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert len(response.json()["data"]) == 6


def test_collection_username(
    client: TestClient, auth_header, mock_user_collection, mock_access_token
):
    mock_access_token(access_token="token", user_id=6, username="ua")
    mock_user_collection(id=1, uid=6, subject_id=1, private=False)
    response = client.get("/v0/user/ua/collections", headers=auth_header)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    assert len(response.json()["data"]) == 1
