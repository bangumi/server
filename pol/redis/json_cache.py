from typing import Any, Dict, Type, Union, TypeVar, Optional

import orjson
from aioredis import Redis
from pydantic import BaseModel, ValidationError
from aioredis.client import KeyT, ExpiryT

__all__ = ["JSONRedis"]

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
        value: Union[Dict[str, Any], int],
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
