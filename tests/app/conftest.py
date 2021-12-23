import pytest
from starlette.testclient import TestClient

import pol.server

access_token = "a_development_access_token"


@pytest.fixture()
def client():
    with TestClient(pol.server.app) as test_client:
        yield test_client
    pol.server.app.dependency_overrides = {}


@pytest.fixture()
def auth_header():
    return {"Authorization": f"Bearer {access_token}"}
