from typing import Dict, Iterator, Protocol
from unittest import mock

import pytest
from fastapi import FastAPI

from pol.models import User, Avatar, PublicUser
from pol.api.v0.depends.auth import Guest
from pol.services.user_service import UserService
from pol.services.subject_service import SubjectService


class MockUserService:
    token: str

    def __init__(self, token: str):
        self.token = token

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

    async def get_by_access_token(self, access_token: str):
        if access_token == self.token:
            # 构造测试账号
            return User(
                id=382951,
                group_id=10,
                username=382951,
                nickname="nickname 382951",
                registration_date=1512262276,
                sign="",
                avatar=Avatar.from_db_record(""),
            )
        return Guest()


@pytest.fixture()
def mock_user_service(app: FastAPI, access_token: str):
    service = mock.Mock(wraps=MockUserService(access_token))

    async def mocker():
        return service

    app.dependency_overrides[UserService.new] = mocker
    yield service
    app.dependency_overrides.pop(UserService.new, None)


class MockSubjectService(Protocol):
    get_by_id: mock.AsyncMock
    get_by_ids: mock.AsyncMock


@pytest.fixture()
def mock_subject_service(app) -> MockSubjectService:
    """
    mock SubjectService, also override dependency `SubjectService.new` for all router
    """
    m: MockSubjectService = mock.Mock()
    m.get_by_id = mock.AsyncMock(return_value=None)
    m.get_by_ids = mock.AsyncMock(return_value=None)

    async def mocker():
        return m

    app.dependency_overrides[SubjectService.new] = mocker
    yield m
    app.dependency_overrides.pop(SubjectService.new, None)
