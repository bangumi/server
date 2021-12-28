import time
import datetime

from loguru import logger
from fastapi import FastAPI
from starlette.requests import Request

from pol import config
from pol.res import ORJSONResponse


def setup_http_middleware(app: FastAPI):
    @app.middleware("http")
    async def add_extra_headers(request: Request, call_next):
        start_time = time.time()
        response = await call_next(request)
        process_time = time.time() - start_time
        response.headers["X-Process-Time"] = str(int(process_time * 1000)) + "ms"
        response.headers["x-server-version"] = config.COMMIT_REF
        return response

    @app.middleware("http")
    async def log(request: Request, call_next):
        try:
            return await call_next(request)
        except Exception as exc:
            ray = request.headers.get("cf-ray")
            logger.exception(
                "catch exception in middleware",
                extra={
                    "url": str(request.url),
                    "query": dict(request.query_params),
                    "x-request-id": request.headers.get("xf-ray", ""),
                    "event": "http.exception",
                    "exception": "{}.{}".format(
                        getattr(exc, "__module__", "builtin"),
                        exc.__class__.__name__,
                    ),
                },
            )
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
                        "time": datetime.datetime.now().astimezone().isoformat(),
                        "report-to": "https://github.com/bangumi/server/issues",
                    },
                },
                status_code=500,
            )
