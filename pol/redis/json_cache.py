from typing import Any, Dict, Union

import orjson
from aioredis import Redis
from aioredis.client import KeyT, ExpiryT


class JSONRedis(Redis):
    async def get(self, name: KeyT) -> Any:
        value = await super().get(name)
        if value is not None:
            try:
                return orjson.loads(value)
            except orjson.JSONDecodeError:
                await self.delete(name)
        return None

    async def set_json(
        self,
        name: KeyT,
        value: Union[Dict[str, Any], int],
        ex: ExpiryT = None,
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
