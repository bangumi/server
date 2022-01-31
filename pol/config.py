import os.path
from pathlib import Path

from starlette.config import Config

PROJ_ROOT = Path(os.path.normpath(os.path.join(os.path.dirname(__file__), "..")))
_config = Config(PROJ_ROOT / "env" / "dev")

APP_NAME = "new bangumi api server"

DEBUG = _config("DEBUG", cast=bool, default=False)

COMMIT_REF = _config("COMMIT_REF", default="dev")

MYSQL_HOST = _config("MYSQL_HOST", default="127.0.0.1")
MYSQL_PORT = _config("MYSQL_PORT", default=3306, cast=int)
MYSQL_USER = _config("MYSQL_USER", default="user")
MYSQL_PASS = _config("MYSQL_PASS", default="password")
MYSQL_DB = _config("MYSQL_DB", default="bangumi")

REDIS_URI = _config("REDIS_URI", default="redis://127.0.0.1:6379/0")

CACHE_KEY_PREFIX = "api-cache:"
