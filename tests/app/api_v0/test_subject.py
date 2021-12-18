from starlette.testclient import TestClient


def test_subject_not_found(client: TestClient):
    response = client.get("/v0/subjects/2000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_subject_not_valid(client: TestClient):
    response = client.get("/v0/subjects/hello")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"

    response = client.get("/v0/subjects/0")
    assert response.status_code == 422
    assert response.headers["content-type"] == "application/json"


def test_subject_basic(client: TestClient):
    response = client.get("/v0/subjects/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["id"] == 2
    assert data["name"] == "坟场"


def test_subject_locked(client: TestClient):
    response = client.get("/v0/subjects/2")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()
    assert data["locked"]


def test_subject_redirect(client: TestClient):
    response = client.get("/v0/subjects/18", allow_redirects=False)
    assert response.status_code == 307
    assert response.headers["location"] == "/v0/subjects/19"


def test_subject_ep_query_limit(client: TestClient):
    response = client.get("/v0/subjects/8/eps", params={"limit": 5})
    assert response.status_code == 200

    data = response.json()
    assert isinstance(data, list)
    assert len(data) == 5

    ids = [x["id"] for x in data]

    new_data = client.get("/v0/subjects/8/eps", params={"limit": 4, "offset": 1}).json()

    assert ids[1:] == [x["id"] for x in new_data]


def test_subject_ep_type(client: TestClient):
    response = client.get("/v0/subjects/253/eps", params={"type": 3})
    assert response.status_code == 200

    data = response.json()
    assert [x["id"] for x in data] == [103233, 103234, 103235]


def test_subject_characters(client: TestClient):
    response = client.get("/v0/subjects/8/characters")
    assert response.status_code == 200

    data = response.json()
    assert isinstance(data, list)
    assert data


def test_subject_persons(client: TestClient):
    response = client.get("/v0/subjects/4/persons")
    assert response.status_code == 200

    data = response.json()

    assert isinstance(data, list)
    assert data


def test_subject_subjects_ban(client: TestClient):
    response = client.get("/v0/subjects/5/subjects")
    assert response.status_code == 404


def test_subject_subjects(client: TestClient):
    response = client.get("/v0/subjects/11/subjects")
    assert response.status_code == 200
    data = response.json()

    assert isinstance(data, list)
    assert data
