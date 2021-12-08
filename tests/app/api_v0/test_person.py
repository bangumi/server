from starlette.testclient import TestClient


def test_person_not_found(client: TestClient):
    response = client.get("/api/v0/person/2000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_person_not_valid(client: TestClient):
    response = client.get("/api/v0/person/hello")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"

    response = client.get("/api/v0/person/0")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"


def test_person_basic(client: TestClient):
    response = client.get("/api/v0/person/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["img"] is None
    assert not data["locked"]


def test_person_subjects(client: TestClient):
    response = client.get("/api/v0/person/1/subjects")
    assert response.status_code == 200

    subjects = response.json()
    assert set(subjects[0].keys()) == {
        "id",
        "staff",
        "name",
        "name_cn",
        "image",
    }


def test_person_redirect(client: TestClient):
    response = client.get("/api/v0/person/10", allow_redirects=False)
    assert response.status_code == 307


def test_person_lock(client: TestClient):
    response = client.get("/api/v0/person/9")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["locked"]
