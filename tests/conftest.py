import pytest
from aioresponses import aioresponses

from pol.db.mysql import Session


def pytest_sessionstart(session):
    """
    Called after the Session object has been created and
    before performing collection and entering the run test loop.
    """
    "session start"


@pytest.fixture()
def mock_aiohttp():
    with aioresponses() as m:
        yield m


@pytest.fixture()
def db_session():
    db_session = Session()
    try:
        yield db_session
    except Exception:
        db_session.rollback()
        raise
    finally:
        db_session.close()
