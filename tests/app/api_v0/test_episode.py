from starlette.testclient import TestClient

from pol.db.const import EpType


def test_episodes(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 8})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"


def test_episodes_offset(client: TestClient):
    response = client.get("/v0/episodes", params={"subject_id": 8, "offset": 3})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    for ep in response.json()["data"]:
        assert ep["sort"] == ep["ep"]


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
