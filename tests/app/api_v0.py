from starlette.testclient import TestClient


def test_person_basic(client: TestClient):
    response = client.get("/api/v0/person/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["img"] is None
    assert isinstance(data["subjects"], list)
    assert not data["lock"]


def test_person_redirect(client: TestClient):
    response = client.get("/api/v0/person/10", allow_redirects=False)
    assert response.status_code == 302


def test_person_lock(client: TestClient):
    response = client.get("/api/v0/person/9")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["lock"]
