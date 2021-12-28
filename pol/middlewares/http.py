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
