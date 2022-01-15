import datetime

import pytest
from sqlalchemy.orm import Session

from pol.db import sa
from pol.models import Permission
from tests.base import async_test
from pol.db.tables import ChiiMember, ChiiUserGroup
from tests.conftest import MockUser, MockAsyncSession
from pol.services.user_service import UserService


@pytest.mark.env("database")
@async_test
async def test_user_service_expired_token(mock_user: MockUser, AsyncSessionMaker):
    mock_user(
        access_token="ttt",
        user_id=200,
        expires=datetime.datetime.now() - datetime.timedelta(days=3),
    )

    async with AsyncSessionMaker() as db:
        with pytest.raises(UserService.NotFoundError):
            await UserService(db).get_by_access_token("ttt")


@pytest.mark.env("database")
@async_test
async def test_user_service__missing_user(
    mock_user: MockUser, db_session: Session, AsyncSessionMaker
):
    mock_user(
        access_token="ttt",
        user_id=200,
        expires=datetime.datetime.now() + datetime.timedelta(days=3),
    )
    db_session.execute(sa.delete(ChiiMember).where(ChiiMember.uid == 200))
    db_session.commit()

    async with AsyncSessionMaker() as db:
        with pytest.raises(UserService.NotFoundError):
            await UserService(db).get_by_access_token("ttt")


@async_test
async def test_user_service_read_permission(mock_db: MockAsyncSession):
    mock_db.get.return_value = ChiiUserGroup(usr_grp_perm={"subject_cover_lock": 1})
    v = await UserService(mock_db).get_permission(180)
    assert v.subject_cover_lock == 1
    assert UserService.cache[180] == v


@async_test
async def test_user_service_permission_fallback(mock_db: MockAsyncSession):
    mock_db.get.return_value = None
    assert Permission() == await UserService(mock_db).get_permission(181)
