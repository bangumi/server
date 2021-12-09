import os
import threading

from loguru import logger
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from fastapi.exceptions import RequestValidationError
from starlette.middleware import cors

from pol import api, res, config
from pol.res import ORJSONResponse
from pol.db.mysql import database
from pol.middlewares.http import setup_http_middleware

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

app.add_middleware(cors.CORSMiddleware, allow_origins=["bgm.tv", "bangumi.tv"])
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


@app.on_event("startup")
async def startup() -> None:
    await database.connect()
    app.state.db = database

    logger.info("server start at pid {}, tid {}", os.getpid(), threading.get_ident())


@app.on_event("shutdown")
async def shutdown() -> None:
    await app.state.db.disconnect()


@app.get("/v0/openapi.json", include_in_schema=False)
async def openapi():
    return app.openapi()


@app.get("/v0", response_class=HTMLResponse, include_in_schema=False)
async def doc():
    return """<!DOCTYPE html>
<html lang=zh-cmn-Hans>
<head>
<link type="text/css" rel="stylesheet"
    href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
<link rel="shortcut icon" href="https://bgm.tv/img/favicon.ico">
<title>Pol server - Swagger UI</title>
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
