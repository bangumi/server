from typing import Dict, Iterator
from unittest import mock

import pytest
from fastapi import FastAPI

from pol.models import Avatar, PublicUser
from pol.services.user_service import UserService

__all__ = ["MockUserService", "mock_user_service"]


class MockUserService:
    async def get_users_by_id(self, ids: Iterator[int]) -> Dict[int, PublicUser]:
        ids = list(ids)
        for uid in ids:
            assert uid > 0

        return {
            uid: PublicUser(
                id=uid,
                username=f"username {uid}",
                nickname=f"nickname {uid}",
                avatar=Avatar.from_db_record(""),
            )
            for uid in ids
        }

    async def get_by_uid(self, uid: int) -> PublicUser:
        assert uid > 0
        return PublicUser(
            id=uid,
            username=f"username {uid}",
            nickname=f"nickname {uid}",
            avatar=Avatar.from_db_record(""),
        )


@pytest.fixture()
def mock_user_service(app: FastAPI):
    service = mock.Mock(wraps=MockUserService())

    async def mocker():
        return service

    app.dependency_overrides[UserService.new] = mocker
    yield service
    app.dependency_overrides.pop(UserService.new, None)
