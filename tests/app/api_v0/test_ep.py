from starlette.testclient import TestClient


def test_episode_404(client: TestClient):
    response = client.get("/v0/episodes/10000000")
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_episode(client: TestClient):
    response = client.get("/v0/episodes/103234")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()

    assert "id" in data
    assert "subject_id" in data
    assert "name" in data


def test_episodes(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 253})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    data = response.json()["data"]

    for ep in data:
        if ep["type"] != 0:
            assert ep["ep"] is None


def test_episodes_404(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 1000000})
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"
