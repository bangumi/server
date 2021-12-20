from starlette.testclient import TestClient


def test_doc_html(client: TestClient):
    response = client.get("/v0", allow_redirects=False)
    assert response.status_code == 200
    assert "text/html" in response.headers["content-type"]


def test_doc_html_2(client: TestClient):
    response = client.get("/v0/", allow_redirects=False)
    assert response.status_code == 200
    assert "text/html" in response.headers["content-type"]


def test_openapi_json(client: TestClient):
    response = client.get("/v0/openapi.json")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


def test_default_404(client: TestClient):
    response = client.get("/non-exist-page")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res == {
        "title": "Not Found",
        "description": "The path you requested doesn't exist",
        "detail": (
            "This is default 404 response, "
            "if you see this response, please check your request path"
        ),
    }
