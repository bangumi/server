import datetime

import pytest
from sqlalchemy.orm import Session

from pol.db import sa
from pol.db.tables import ChiiMember
from tests.conftest import MockUser
from pol.services.user_service import UserService


@pytest.mark.env("database")
@pytest.mark.asyncio()
async def test_auth_expired_token(mock_user: MockUser, AsyncSessionMaker):
    mock_user(
        access_token="ttt",
        user_id=200,
        expires=datetime.datetime.now() - datetime.timedelta(days=3),
    )

    async with AsyncSessionMaker() as db:
        with pytest.raises(UserService.NotFoundError):
            await UserService(db).get_by_access_token("ttt")


@pytest.mark.env("database")
@pytest.mark.asyncio()
async def test_auth_missing_user(
    mock_user: MockUser,
    db_session: Session,
    AsyncSessionMaker,
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
