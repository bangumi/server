import os
import datetime
import threading

import pydantic
import pymysql.err  # type: ignore
import sqlalchemy.exc
from loguru import logger
from fastapi import FastAPI
from databases import DatabaseURL
from sqlalchemy.orm import sessionmaker
from fastapi.responses import HTMLResponse
from fastapi.exceptions import RequestValidationError
from starlette.requests import Request
from starlette.exceptions import HTTPException as StarletteHTTPException
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine

from pol import api, res, config
from pol.res import ORJSONResponse
from pol.db.mysql import database
from pol.middlewares.http import setup_http_middleware
from pol.redis.json_cache import JSONRedis

app = FastAPI(
    debug=config.DEBUG,
    title=config.APP_NAME,
    version=config.COMMIT_REF,
    docs_url=None,
    redoc_url=None,
    swagger_ui_oauth2_redirect_url=None,
    openapi_url=None,
    default_response_class=ORJSONResponse,
)

setup_http_middleware(app)
app.include_router(api.router)


@app.exception_handler(res.HTTPException)
async def http_exception_handler(request, exc: res.HTTPException):
    return ORJSONResponse(
        {
            "title": exc.title,
            "description": exc.description,
            "detail": exc.detail,
        },
        headers=exc.headers,
        status_code=exc.status_code,
    )


@app.exception_handler(StarletteHTTPException)
async def global_404(request, exc: StarletteHTTPException):
    if exc.status_code == 404:
        return ORJSONResponse(
            {
                "title": "Not Found",
                "description": "The path you requested doesn't exist",
                "detail": (
                    "This is default 404 response, "
                    "if you see this response, please check your request path"
                ),
            },
            status_code=exc.status_code,
        )

    return ORJSONResponse({"detail": exc.detail}, status_code=exc.status_code)


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request, exc: RequestValidationError):
    return ORJSONResponse(
        {
            "title": "Invalid Request",
            "description": "One or more parameters to your request was invalid.",
            "detail": exc.errors(),
        },
        status_code=422,
    )


@app.exception_handler(sqlalchemy.exc.SQLAlchemyError)
@app.exception_handler(pymysql.err.MySQLError)
async def global_pymysql_error(
    request: Request, exc: pymysql.err.ProgrammingError
):  # pragma: no cover
    ray = request.headers.get("cf_ray")
    logger.exception("exception in sqlalchemy", cf_ray=ray)
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


@app.exception_handler(pydantic.ValidationError)
async def internal_response_validation_error(
    request, exc: pydantic.ValidationError
):  # pragma: no cover
    """this is intend to cache un-expected pydantic validation error"""
    ray = request.headers.get("cf_ray")
    logger.exception(str(exc), message="internal validation error", cf_ray=ray)
    return ORJSONResponse(
        {
            "title": "Internal Server Error",
            "description": (
                "something unexpected happened, please report to maintainer"
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


@app.on_event("startup")
async def startup() -> None:
    await database.connect()
    app.state.db = database
    app.state.redis = await JSONRedis.from_url(config.REDIS_URI)
    app.state.engine = engine = create_async_engine(
        str(DatabaseURL(config.MYSQL_URI).replace(dialect="mysql+asyncmy")),
        echo=True,
    )

    app.state.Session = sessionmaker(
        engine, expire_on_commit=False, class_=AsyncSession
    )

    logger.info("server start at pid {}, tid {}", os.getpid(), threading.get_ident())


@app.on_event("shutdown")
async def shutdown() -> None:
    await app.state.db.disconnect()
    await app.state.redis.close()
    await app.state.engine.dispose()


@app.get("/v0/openapi.json", include_in_schema=False)
async def openapi():
    return app.openapi()


@app.get("/v0/", response_class=HTMLResponse, include_in_schema=False)
@app.get("/v0", response_class=HTMLResponse, include_in_schema=False)
async def doc():
    return """<!DOCTYPE html>
<html lang=zh-cmn-Hans>
<head>
<link type="text/css" rel="stylesheet"
    href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
<link rel="shortcut icon" href="https://bgm.tv/img/favicon.ico">
<title>Bangumi API - Swagger UI</title>
<style>
    a {
        text-decoration: none
    }
</style>
</head>
<body>
<div id="swagger-ui"></div>
<hr>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js">
</script>
<script>
const ui = SwaggerUIBundle({
    url: '/v0/openapi.json',
    dom_id: '#swagger-ui',
    presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
    ],
    layout: "BaseLayout",
    deepLinking: true
})
</script>
</body>
</html>
"""
