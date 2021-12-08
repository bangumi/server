import pytest
from starlette.testclient import TestClient

import pol.server


@pytest.fixture()
def client():
    with TestClient(pol.server.app) as test_client:
        yield test_client
    pol.server.app.dependency_overrides = {}
