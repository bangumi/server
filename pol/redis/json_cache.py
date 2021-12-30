import functools
from abc import abstractmethod
from typing import (
    Any,
    Dict,
    Type,
    Union,
    Generic,
    TypeVar,
    Callable,
    Optional,
    Protocol,
    Awaitable,
    cast,
)

import orjson
from aioredis import Redis
from pydantic import BaseModel, ValidationError
from aioredis.client import KeyT, ExpiryT
from starlette.responses import Response

from pol.config import CACHE_KEY_PREFIX

DataType = Union[Dict[str, Any], int]

T = TypeVar("T", bound=BaseModel)


class JSONRedis(Redis):
    async def get(self, name: KeyT) -> Any:
        value = await super().get(name)
        if value is not None:
            try:
                return orjson.loads(value)
            except orjson.JSONDecodeError:
                await self.delete(name)
        return None

    async def get_with_model(self, name: KeyT, model: Type[T]) -> Optional[T]:
        """will also try to parse cached json as a pydantic model.
        cache will be purged if it's broken
        """
        value = await super().get(name)
        if value is not None:
            try:
                return model.parse_obj(orjson.loads(value))
            except (orjson.JSONDecodeError, ValidationError):
                await self.delete(name)
        return None

    async def set_json(
        self,
        name: KeyT,
        value: DataType,
        ex: ExpiryT,
        px: ExpiryT = None,
        nx: bool = False,
        xx: bool = False,
        keepttl: bool = False,
    ):
        return await self.set(
            name=name,
            value=orjson.dumps(value),
            ex=ex,
            px=px,
            nx=nx,
            xx=xx,
            keepttl=keepttl,
        )


class KeyBuilder(Protocol):
    @abstractmethod
    def __call__(self, **kwargs) -> str:
        pass


T1 = TypeVar("T1", bound=DataType)


class APIHandler(Generic[T1]):
    @abstractmethod
    def __call__(
        self,
        *args: Any,
        response: Response = None,
        redis: JSONRedis = None,
        **kwargs: Any,
    ) -> Awaitable[T1]:
        pass


def cache(
    key_builder: KeyBuilder,
    ex: int = 60,
    on_validate_cache: Callable[[T1], bool] = None,
):
    def wrapper(func: APIHandler[T1]):
        @functools.wraps(func)
        async def inner(*args, **kwargs) -> T1:
            response: Response = kwargs["response"]
            redis: JSONRedis = kwargs["redis"]
            cache_key = f"{CACHE_KEY_PREFIX}{key_builder(**kwargs)}"
            if can_use_cache := response is not None and redis is not None:
                is_valid = False
                value: Optional[T1] = None
                try:
                    value = await redis.get(cache_key)
                    if value is not None:
                        is_valid = (
                            on_validate_cache(value)
                            if on_validate_cache is not None
                            else True
                        )
                except (orjson.JSONDecodeError, ValidationError):
                    pass
                if not is_valid:
                    if value is not None:
                        await redis.delete(cache_key)
                    response.headers["x-cache-status"] = "miss"
                else:
                    response.headers["x-cache-status"] = "hit"
                    return cast(T1, value)
            data = await func(*args, **kwargs)
            if can_use_cache:
                await redis.set_json(cache_key, data, ex=ex)
            return data

        return inner

    return wrapper
