from unittest import mock

import pytest

from pol import res
from tests.base import async_test
from pol.api.v0.depends.auth import get_current_user
from pol.services.user_service import UserService


@async_test
async def test_depends_get_current_user(mock_redis):
    mock_user_service = mock.Mock()
    mock_user_service.get_by_access_token = mock.AsyncMock(
        side_effect=UserService.NotFoundError
    )

    with pytest.raises(res.HTTPException) as exc:
        await get_current_user(
            token="access token", service=mock_user_service, redis=mock_redis
        )

    assert exc.value.status_code == 403
