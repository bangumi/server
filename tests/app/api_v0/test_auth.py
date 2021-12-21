from starlette.testclient import TestClient


def test_auth_401(client: TestClient):
    response = client.get("/v0/episodes/10000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"
