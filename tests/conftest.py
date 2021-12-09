import pytest
from databases import DatabaseURL
from sqlalchemy import create_engine
from aioresponses import aioresponses
from sqlalchemy.orm import sessionmaker

from pol import config


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


engine = create_engine(
    str(DatabaseURL(config.MYSQL_URI).replace(dialect="mysql+pymysql"))
)

Session = sessionmaker(bind=engine)


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
