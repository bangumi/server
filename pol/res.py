from typing import Any, Dict, Type, Optional

import orjson
from starlette.responses import JSONResponse


class ORJSONResponse(JSONResponse):
    media_type = "application/json"

    def render(self, content: Any) -> bytes:
        return orjson.dumps(content)


class HTTPException(Exception):
    def __init__(
        self,
        status_code: int,
        title: str,
        description: str,
        detail: Any = None,
        headers: Optional[Dict[str, Any]] = None,
    ) -> None:
        self.status_code = status_code
        self.detail = detail
        self.headers = headers
        self.title = title
        self.description = description


def response(
    model: Type = None, description: str = None, headers=None
) -> Dict[str, Any]:
    d: Dict[str, Any] = {}
    if model is not None:
        d["model"] = model
    if description:
        d["description"] = description
    if headers is not None:
        d["headers"] = headers
    return d
