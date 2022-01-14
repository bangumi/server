from typing import Callable
from datetime import datetime

import pymysql.err  # type: ignore
import sqlalchemy.exc
from loguru import logger
from fastapi import Request, Response
from fastapi.routing import APIRoute
from fastapi.exceptions import RequestValidationError

from pol import res
from pol.res import ORJSONResponse


class ErrorCatchRoute(APIRoute):
    """starlette不支持全局 catch `Exception`，只能用这种办法来捕获内部异常。"""

    def get_route_handler(self) -> Callable:
        original_route_handler = super().get_route_handler()

        async def custom_route_handler(request: Request) -> Response:
            try:
                return await original_route_handler(request)
            except RequestValidationError as exc:
                return ORJSONResponse(
                    {
                        "title": "Invalid Request",
                        "description": "One or more parameters are not valid.",
                        "detail": exc.errors(),
                    },
                    status_code=422,
                )
            except res.HTTPException as exc:
                return ORJSONResponse(
                    {
                        "title": exc.title,
                        "description": exc.description,
                        "detail": exc.detail,
                    },
                    headers=exc.headers,
                    status_code=exc.status_code,
                )
            except (
                pymysql.err.MySQLError,
                sqlalchemy.exc.SQLAlchemyError,
            ) as exc:  # pragma: no cover
                ray = request.headers.get("cf-ray")
                logger.exception("exception in sqlalchemy {}", str(exc), cf_ray=ray)
                return ORJSONResponse(
                    {
                        "title": "Internal Server Error",
                        "description": (
                            "something unexpected happened with mysql,"
                            " please report to maintainer"
                        ),
                        "detail": {
                            "cf-ray": ray,
                            "url": str(request.url),
                            "time": datetime.now().astimezone().isoformat(),
                            "report-to": "https://github.com/bangumi/server/issues",
                        },
                    },
                    status_code=500,
                )
            except Exception:  # pragma: no cover
                ray = request.headers.get("cf-ray")
                logger.exception(
                    "unexpected exception {} {}",
                    type(Exception),
                    str(Exception),
                    cf_ray=ray,
                )
                return ORJSONResponse(
                    {
                        "title": "Internal Server Error",
                        "description": (
                            "something unexpected happened, please report to maintainer"
                        ),
                        "detail": {
                            "cf-ray": ray,
                            "url": str(request.url),
                            "time": datetime.now().astimezone().isoformat(),
                            "report-to": "https://github.com/bangumi/server/issues",
                        },
                    },
                    status_code=500,
                )

        return custom_route_handler
