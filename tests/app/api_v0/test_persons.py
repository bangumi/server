from sqlalchemy.orm import Session
from starlette.testclient import TestClient

from pol.db.tables import ChiiPerson


def test_persons_basic_router(client: TestClient):
    response = client.get("/api/v0/persons")
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res


def test_persons_filter_name(client: TestClient):
    name = "和田薫"
    response = client.get("/api/v0/persons", params={"name": name})
    assert response.status_code == 200
    assert response.headers["content-type"] == "application/json"

    res = response.json()
    assert res["total"]
    assert res["offset"] == 0
    assert "limit" in res
    assert res["data"]
    assert res["data"][0]["name"] == name


def test_persons_filter_career_valid(client: TestClient):
    career = "artist"
    response = client.get("/api/v0/persons", params={"career": career})
    assert response.status_code == 200
    res = response.json()

    assert [x for x in res["data"] if career in x["career"]]


def test_persons_filter_career_invalid(client: TestClient):
    response = client.get("/api/v0/persons", params={"career": "e"})
    assert response.status_code == 422


def test_persons_page_limit(client: TestClient):
    limit = 3
    response = client.get("/api/v0/persons", params={"limit": limit})
    assert response.status_code == 200

    res = response.json()
    assert len(res["data"]) == limit


def test_persons_page_limit_too_big(client: TestClient, db_session: Session):
    limit = 30000
    response = client.get("/api/v0/persons", params={"limit": limit})
    assert response.status_code == 422, response.text


def test_persons_page_limit_big_omit(client: TestClient, db_session: Session):
    limit = 50

    response = client.get("/api/v0/persons", params={"limit": limit})
    assert response.status_code == 200, response.text

    res = response.json()
    assert (
        len(res["data"])
        == db_session.query(ChiiPerson)
        .filter(ChiiPerson.prsn_ban == 0, ChiiPerson.prsn_redirect == 0)
        .count()
    )


def test_persons_page_sort_invalid(client: TestClient, db_session: Session):
    response = client.get("/api/v0/persons", params={"sort": "s"})
    assert response.status_code == 422, response.text


def test_persons_page_offset(client: TestClient, db_session: Session):
    response = client.get("/api/v0/persons", params={"offset": 1})
    assert response.status_code == 200, response.text

    expected = [
        x.prsn_id
        for x in db_session.query(ChiiPerson.prsn_id)
        .filter(ChiiPerson.prsn_ban == 0, ChiiPerson.prsn_redirect == 0)
        .offset(1)
        .order_by(ChiiPerson.prsn_id.desc())
    ]

    res = response.json()
    assert res["offset"] == 1
    assert [x["id"] for x in res["data"]] == expected


def test_persons_sort_valid(client: TestClient, db_session: Session):
    response = client.get("/api/v0/persons", params={"sort": "id"})
    assert response.status_code == 200, response.text

    expected = [
        x.prsn_id
        for x in db_session.query(ChiiPerson.prsn_id)
        .filter(ChiiPerson.prsn_ban == 0, ChiiPerson.prsn_redirect == 0)
        .order_by(ChiiPerson.prsn_id.desc())
    ]

    res = response.json()
    assert [x["id"] for x in res["data"]] == expected
