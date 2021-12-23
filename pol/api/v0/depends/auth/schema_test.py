import pytest
from starlette.requests import Request
from starlette.datastructures import Headers

from pol import res
from tests.base import async_test
from pol.api.v0.depends.auth import API_KEY_HEADER, OPTIONAL_API_KEY_HEADER


def req(s: str = None):
    if s:
        header = Headers({"Authorization": s}).raw
    else:
        header = []
    return Request(scope={"type": "http", "headers": header})


@async_test
async def test_schema_optional_non_header():
    await OPTIONAL_API_KEY_HEADER(req())


@async_test
async def test_schema_optional_1():
    assert "a development access token" == await OPTIONAL_API_KEY_HEADER(
        req("bearer a development access token")
    )


@async_test
async def test_schema_optional_2():
    with pytest.raises(res.HTTPException) as exc_info:
        await OPTIONAL_API_KEY_HEADER(req("vv"))
    assert exc_info.value.status_code == 403


@async_test
async def test_schema_optional_3():
    with pytest.raises(res.HTTPException) as exc_info:
        await OPTIONAL_API_KEY_HEADER(req("vv bb"))
    assert exc_info.value.status_code == 403


@async_test
async def test_schema_required_1():
    with pytest.raises(res.HTTPException) as exc_info:
        await API_KEY_HEADER(req())
    assert exc_info.value.status_code == 403


@async_test
async def test_schema_required_2():
    with pytest.raises(res.HTTPException) as exc_info:
        await API_KEY_HEADER(req("vv bb"))
    assert exc_info.value.status_code == 403


@async_test
async def test_schema_required_3():
    assert "a development access token" == await API_KEY_HEADER(
        req("bearer a development access token")
    )


@async_test
async def test_schema_required_4():
    assert "a-development_access-token" == await API_KEY_HEADER(
        req("bearer a-development_access-token")
    )
