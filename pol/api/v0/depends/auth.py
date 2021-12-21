from typing import Optional
from datetime import datetime, timedelta

from fastapi import Depends
from pydantic import Field, BaseModel, ValidationError
from databases import Database
from starlette.status import HTTP_403_FORBIDDEN
from starlette.requests import Request
from fastapi.security.http import SecurityBase
from fastapi.openapi.models import HTTPBearer as HTTPBearerModel
from fastapi.security.utils import get_authorization_scheme_param

from pol import res, curd, config
from pol.curd import NotFoundError
from pol.depends import get_db, get_redis
from pol.db.tables import ChiiMember, ChiiOauthAccessToken
from pol.permission import UserGroup
from pol.redis.json_cache import JSONRedis


class OptionalHTTPBearer(SecurityBase):
    def __init__(
        self,
        *,
        bearerFormat: Optional[str] = None,
        description: Optional[str] = (
            "Some api like getting a subject doesn't require any authorization."
            "But only authorized user can see NSFW recoed, "
            "non-authorized user get 404 response"
        ),
    ):
        self.model = HTTPBearerModel(bearerFormat=bearerFormat, description=description)
        self.scheme_name = self.__class__.__name__

    async def __call__(self, request: Request) -> str:
        authorization: str = request.headers.get("Authorization")
        if not authorization:
            return ""
        scheme, credentials = get_authorization_scheme_param(authorization)
        if not (authorization and scheme and credentials):
            raise res.HTTPException(
                title="unauthorized",
                status_code=HTTP_403_FORBIDDEN,
                description="Not authenticated",
            )
        if scheme.lower() != "bearer":
            raise res.HTTPException(
                status_code=HTTP_403_FORBIDDEN,
                title="unauthorized",
                description="Invalid authentication credentials",
            )
        return credentials


class HTTPBearer(SecurityBase):
    def __init__(
        self,
        *,
        bearerFormat: Optional[str] = None,
        description: Optional[str] = None,
    ):
        self.model = HTTPBearerModel(bearerFormat=bearerFormat, description=description)
        self.scheme_name = self.__class__.__name__

    async def __call__(self, request: Request) -> str:
        authorization: str = request.headers.get("Authorization")
        scheme, credentials = get_authorization_scheme_param(authorization)
        if not (authorization and scheme and credentials):
            raise res.HTTPException(
                title="unauthorized",
                status_code=HTTP_403_FORBIDDEN,
                description="Not authenticated",
            )
        if scheme.lower() != "bearer":
            raise res.HTTPException(
                status_code=HTTP_403_FORBIDDEN,
                title="unauthorized",
                description="Invalid authentication credentials",
            )
        return credentials


API_KEY_HEADER = HTTPBearer()
OPTIONAL_API_KEY_HEADER = OptionalHTTPBearer()


class User(BaseModel):
    id: int
    username: str
    nickname: str
    group_id: UserGroup = Field(alias="groupid")
    registration_date: datetime = Field(alias="regdate")

    # lastvisit: int
    # lastactivity: int
    # lastpost: int
    # dateformat: str
    # timeformat: int
    # timeoffset: str
    # newpm: int
    # new_notify: int
    # sign: str

    def allow_nsfw(self) -> bool:
        allow_date = self.registration_date + timedelta(days=60)
        return datetime.utcnow().astimezone() > allow_date


async def optional_user(
    token: str = Depends(OPTIONAL_API_KEY_HEADER),
    db: Database = Depends(get_db),
    redis: JSONRedis = Depends(get_redis),
) -> Optional[User]:
    cache_key = config.CACHE_KEY_PREFIX + f"access:{token}"
    if value := await redis.get(cache_key):
        try:
            return User.parse_obj(value)
        except ValidationError:
            await redis.delete(cache_key)

    try:
        access_row = await curd.get_one(
            db,
            ChiiOauthAccessToken,
            ChiiOauthAccessToken.access_token == token,
            ChiiOauthAccessToken.expires > datetime.now(),
        )

        member_row = await curd.get_one(
            db,
            ChiiMember,
            ChiiMember.uid == int(access_row.user_id),
        )
    except NotFoundError:
        return

    user = User(
        id=member_row.uid,
        groupid=member_row.groupid,
        username=member_row.username,
        nickname=member_row.nickname,
        regdate=member_row.regdate,
    )

    await redis.set_json(cache_key, user.dict(by_alias=True))

    return user


async def get_current_user(
    user: Optional[User] = Depends(optional_user),
) -> User:
    if user is None:
        raise res.HTTPException(
            status_code=HTTP_403_FORBIDDEN,
            title="unauthorized",
            description="Invalid authentication credentials",
        )
    return user
