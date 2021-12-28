import os
import threading

from loguru import logger
from fastapi import FastAPI
from sqlalchemy.orm import sessionmaker
from fastapi.responses import HTMLResponse
from starlette.exceptions import HTTPException as StarletteHTTPException
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine

from pol import api, config
from pol.res import ORJSONResponse
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
    app.state.redis = await JSONRedis.from_url(config.REDIS_URI)
    app.state.engine = engine = create_async_engine(
        "mysql+aiomysql://{}:{}@{}:{}/{}".format(
            config.MYSQL_USER,
            config.MYSQL_PASS,
            config.MYSQL_HOST,
            config.MYSQL_PORT,
            config.MYSQL_DB,
        )
    )
    app.state.Session = sessionmaker(
        engine, expire_on_commit=False, class_=AsyncSession
    )

    logger.info("server start at pid {}, tid {}", os.getpid(), threading.get_ident())


@app.on_event("shutdown")
async def shutdown() -> None:
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
