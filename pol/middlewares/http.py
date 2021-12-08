import time

from fastapi import FastAPI
from starlette.requests import Request

from pol import config


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
            app.state.logger.exception(
                "catch exception in middleware",
                extra={
                    "url": str(request.url),
                    "query": dict(request.query_params),
                    "x-request-id": request.headers.get("x-request-id", ""),
                    "event": "http.exception",
                    "exception": "{}.{}".format(
                        getattr(exc, "__module__", "builtin"),
                        exc.__class__.__name__,
                    ),
                },
            )
            raise
