from operator import itemgetter

from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.db.tables import ChiiCharacter

path = "/v0/characters"


def test_characters_basic_router(client: TestClient):
    response = client.get(path)
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res


def test_characters_filter_name(client: TestClient):
    name = "古河渚"
    response = client.get(path, params={"name": name})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res
    assert res["data"]
    assert res["data"][0]["name"] == name


def test_characters_filter_type(client: TestClient):
    type = 1
    response = client.get("/v0/persons", params={"type": type})
    assert response.status_code == 200

    res = response.json()
    assert res["data"]
    assert [x for x in res["data"] if type == x["type"]]


def test_characters_page_limit(client: TestClient):
    limit = 3
    response = client.get(path, params={"limit": limit})
    assert response.status_code == 200

    res = response.json()
    assert len(res["data"]) == limit


def test_characters_page_limit_too_big(client: TestClient, db_session: Session):
    limit = 30000
    response = client.get(path, params={"limit": limit})
    assert response.status_code == 422, response.text


def test_characters_page_limit_big_omit(client: TestClient, db_session: Session):
    limit = 50

    response = client.get(path, params={"limit": limit})
    assert response.status_code == 200, response.text

    res = response.json()
    assert (
        len(res["data"])
        == db_session.query(ChiiCharacter)
        .filter(ChiiCharacter.crt_ban == 0, ChiiCharacter.crt_redirect == 0)
        .count()
    )


def test_characters_page_sort_invalid(client: TestClient, db_session: Session):
    response = client.get(path, params={"sort": "s"})
    assert response.status_code == 422, response.text


def test_characters_page_offset_too_bid(client: TestClient, db_session: Session):
    response = client.get(path, params={"offset": 100000000})
    assert response.status_code == 422, response.text


def test_characters_page_offset(client: TestClient, db_session: Session):
    response = client.get(path, params={"offset": 1})
    assert response.status_code == 200, response.text

    expected = [
        x.crt_id
        for x in db_session.query(ChiiCharacter.crt_id)
        .filter(ChiiCharacter.crt_ban == 0, ChiiCharacter.crt_redirect == 0)
        .order_by(ChiiCharacter.crt_id.desc())
        .offset(1)
    ]

    res = response.json()
    assert res["offset"] == 1
    assert [x["id"] for x in res["data"]] == expected


def test_characters_sort_valid(client: TestClient, db_session: Session):
    response = client.get(path, params={"sort": "id"})
    assert response.status_code == 200, response.text

    expected = [
        x.crt_id
        for x in db_session.query(ChiiCharacter.crt_id)
        .filter(ChiiCharacter.crt_ban == 0, ChiiCharacter.crt_redirect == 0)
        .order_by(ChiiCharacter.crt_id.desc())
    ]

    res = response.json()
    assert [x["id"] for x in res["data"]] == expected


def test_characters_page_sort_args(client: TestClient, db_session: Session):
    response = client.get(path, params={"sort": "name", "order": 1})
    assert response.status_code == 200, response.text

    res = response.json()
    assert [x["id"] for x in res["data"]] == [
        x["id"] for x in sorted(res["data"], key=itemgetter("name"))
    ]
