from starlette.testclient import TestClient


def test_openapi_json(client: TestClient):
    response = client.get("/api/v0/person/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["img"] is None
    assert isinstance(res["subjects"], list)
