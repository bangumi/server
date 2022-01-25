import pytest
from fastapi import FastAPI
from _pytest.nodes import Node
from starlette.testclient import TestClient

import pol.server

__all__ = ["pytest_addoption",
           "pytest_configure",
           "pytest_runtest_setup",
           "app",
           "client",
           "access_token",
           "auth_header"]

# mockers
pytest_plugins = [
    "tests.fixtures.mock_redis",
    "tests.fixtures.mock_db",
    "tests.fixtures.mock_db_record",
    "tests.fixtures.mock_service",
]


def pytest_addoption(parser):
    parser.addoption("--e2e", action="store_true", help="Run E2E tests")
    parser.addoption(
        "--database", action="store_true", help="Enable tests require database"
    )
    parser.addoption("--redis", action="store_true", help="Enable tests require redis")


def pytest_configure(config):
    # register an additional `pytest.mark.env` marker
    config.addinivalue_line(
        "markers",
        "env(name1, ...name2): mark test to run only on named environment",
    )


def pytest_runtest_setup(item: Node):
    """skip by `@pytest.mark.env` and flag defined in `pytest_addoption`"""
    mark = item.get_closest_marker(name="env")
    if not mark:
        return
    for name in mark.args:
        if not item.config.getoption(name):
            pytest.skip(f"test skipped without flag --{name}")


@pytest.fixture()
def app() -> FastAPI:
    """app instance"""
    return pol.server.app


@pytest.fixture()
def client(app):
    with TestClient(app) as test_client:
        yield test_client
    app.dependency_overrides = {}


@pytest.fixture()
def access_token():
    return "a_development_access_token"


@pytest.fixture()
def auth_header(access_token: str):
    return {"Authorization": f"Bearer {access_token}"}
