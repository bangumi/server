from starlette.testclient import TestClient


def test_doc_html(client: TestClient):
    response = client.get("/v0")
    assert response.status_code == 200
    assert "text/html" in response.headers["content-type"]


def test_openapi_json(client: TestClient):
    response = client.get("/v0/openapi.json")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"
