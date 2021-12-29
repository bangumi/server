from starlette.testclient import TestClient

from pol.db.const import EpType


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
            assert ep["ep"] == 0


def test_episodes_404(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 1000000})
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_episodes_nsfw_non_auth(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 16})
    assert response.status_code == 404
    assert response.headers["content-type"] == "application/json"


def test_episodes_nsfw_auth(client: TestClient, auth_header):
    response = client.get(
        "/v0/episodes",
        params={"subject_id": 16},
        headers=auth_header,
    )
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


def test_episodes_offset(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 8, "offset": 3})
    assert response.status_code == 200

    for ep in response.json()["data"]:
        assert ep["sort"] == ep["ep"]


def test_episodes_non_normal_offset(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 253})
    assert response.status_code == 200

    for ep in response.json()["data"]:
        if ep["type"] != EpType.normal:
            assert ep["ep"] == 0


def test_episodes_start_non_1(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": "211567"})
    assert response.status_code == 200

    res = response.json()

    assert [x["ep"] for x in res["data"]] == list(range(1, 1 + res["total"]))


def test_episodes_start_non_offset(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": "211567", "offset": 3})
    assert response.status_code == 200

    res = response.json()

    assert [x["ep"] for x in res["data"] if x["type"] == EpType.normal] == [
        float(x) for x in range(4, 4 + res["total"] - 3)
    ]

    assert [x["sort"] for x in res["data"] if x["type"] == EpType.normal] == [
        26.0 + i for i in range(19)
    ]
