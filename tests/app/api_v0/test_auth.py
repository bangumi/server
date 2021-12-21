from starlette.testclient import TestClient

access_token = "a_development_access_token"


def test_auth_200(client: TestClient):
    response = client.get("/v0/me", headers={"Authorization": f"Bearer {access_token}"})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


def test_auth_403(client: TestClient):
    response = client.get("/v0/me", headers={"Authorization": "Bearer "})
    assert response.status_code == 403, "no token"

    response = client.get("/v0/me", headers={"Authorization": f"t {access_token}"})
    assert response.status_code == 403, "no token"
