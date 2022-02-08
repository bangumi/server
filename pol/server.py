import os
import threading

import aioredis
from loguru import logger
from fastapi import FastAPI
from sqlalchemy.orm import sessionmaker
from fastapi.responses import HTMLResponse, ORJSONResponse
from starlette.requests import Request
from starlette.responses import Response
from starlette.exceptions import HTTPException as StarletteHTTPException
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine

from pol import api, config
from pol.router import ErrorCatchRoute
from pol.middlewares.http import setup_http_middleware
from pol.redis.json_cache import JSONRedis

app = FastAPI(
    debug=config.DEBUG,
    title=config.APP_NAME,
    version=config.COMMIT_REF,
    docs_url=None,
    redoc_url=None,
    description="你可以在 <https://api.bgm.tv/v0/oauth/> 生成一个 Access Token",
    swagger_ui_oauth2_redirect_url=None,
    openapi_url=None,
    default_response_class=ORJSONResponse,
)

app.router.route_class = ErrorCatchRoute


@app.middleware("http")
async def handle_public_cache(request: Request, call_next):
    """add cache control header, use `pol.http_cache.depends.CacheControl` in handler"""
    response: Response = await call_next(request)
    if v := getattr(request.state, "public_resource", 0):
        response.headers["cache-control"] = f"public, max-age={v}"
    return response


setup_http_middleware(app)
app.include_router(api.router)


@app.exception_handler(StarletteHTTPException)
async def global_404(request, exc: StarletteHTTPException):
    """global 404 handler"""
    if exc.status_code == 404:
        return ORJSONResponse(
            {
                "title": "Not Found",
                "description": "The path you requested doesn't exist",
                "url": "https://api.bgm.tv/v0/",
                "detail": (
                    "This is default 404 response, "
                    "if you see this response, please check your request path"
                ),
            },
            status_code=exc.status_code,
        )

    return ORJSONResponse({"detail": exc.detail}, status_code=exc.status_code)


@app.on_event("startup")
async def startup() -> None:
    app.state.redis_pool = aioredis.ConnectionPool.from_url(config.REDIS_URI)

    app.state.redis = JSONRedis(connection_pool=app.state.redis_pool)
    app.state.engine = engine = create_async_engine(
        "mysql+aiomysql://{}:{}@{}:{}/{}".format(
            config.MYSQL_USER,
            config.MYSQL_PASS,
            config.MYSQL_HOST,
            config.MYSQL_PORT,
            config.MYSQL_DB,
        ),
        pool_recycle=14400,
        pool_size=10,
        max_overflow=20,
    )
    app.state.Session = sessionmaker(
        engine, expire_on_commit=False, class_=AsyncSession
    )

    logger.info("server start at pid {}, tid {}", os.getpid(), threading.get_ident())


@app.on_event("shutdown")
async def shutdown() -> None:
    r: JSONRedis = app.state.redis
    pool: aioredis.ConnectionPool = app.state.redis_pool
    await r.close()
    await pool.disconnect()
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
