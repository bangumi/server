import pytest
from starlette.testclient import TestClient

from tests.base import async_test
from pol.db.tables import ChiiSubject, ChiiSubjectField
from pol.services.subject_service import SubjectService


@pytest.mark.env("database")
@async_test
async def test_get_by_id(client: TestClient, AsyncSessionMaker):
    async with AsyncSessionMaker() as db:
        with pytest.raises(SubjectService.NotFoundError):
            await SubjectService(db).get_by_id(2000000)


@async_test
async def test_get_basic(client: TestClient, AsyncSessionMaker, mock_db):
    mock_db.get.return_value = ChiiSubject(
        subject_id=1,
        subject_type_id=2,
        subject_platform=1,
        subject_name="n",
        subject_name_cn="nc",
        subject_image="",
        field_summary="sm",
        field_infobox="i",
        subject_ban=0,
        fields=ChiiSubjectField(field_redirect=0),
    )
    s = await SubjectService(mock_db).get_by_id(2000000)
    assert s.redirect == 0
    assert s.name == "n"
    assert s.name_cn == "nc"
