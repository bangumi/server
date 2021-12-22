from datetime import datetime

import redis
import pytest
from databases import DatabaseURL
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from pol import sa, config
from pol.db.const import SubjectType
from pol.db.tables import ChiiSubject, ChiiSubjectField


def pytest_sessionstart(session):
    """
    Called after the Session object has been created and
    before performing collection and entering the run test loop.
    """
    "session start"


engine = create_engine(
    str(DatabaseURL(config.MYSQL_URI).replace(dialect="mysql+pymysql"))
)

Session = sessionmaker(bind=engine)


@pytest.fixture()
def db_session():
    db_session = Session()
    try:
        yield db_session
    except Exception:  # pragma: no cover
        db_session.rollback()
        raise
    finally:
        db_session.close()


@pytest.fixture()
def redis_client():
    with redis.Redis.from_url(config.REDIS_URI) as redis_client:
        redis_client.flushdb()
        yield redis_client
        redis_client.flushdb()


@pytest.fixture()
def mock_subject(db_session: Session):
    mock_subject_id = set()

    def mock_id(
        subject_id: int,
        subject_name="",
        subject_type=SubjectType.anime,
        subject_name_cn="",
        subject_uid=1,
        subject_creator=1,
        subject_image="",
        field_infobox="",
        field_summary="",
        field_5="",
        subject_idx_cn="",
        subject_airtime=1,
        subject_nsfw=0,
        subject_ban=0,
        field_tags="",
        field_airtime=1,
        field_year=2007,
        field_mon=1,
        field_week_day=3,
        field_date=datetime.now().astimezone().date(),
    ):
        if db_session.get(ChiiSubject, {"subject_id": subject_id}):  # pragma: no cover
            raise ValueError(f"subject {subject_id} is in database, can't mock it")
        mock_subject_id.add(subject_id)
        db_session.execute(
            sa.delete(ChiiSubject).where(ChiiSubject.subject_id == subject_id)
        )
        db_session.add(
            ChiiSubject(
                subject_id=200,
                subject_name=subject_name,
                subject_name_cn=subject_name_cn,
                subject_uid=subject_uid,
                subject_type_id=subject_type,
                subject_creator=subject_creator,
                subject_image=subject_image,
                field_infobox=field_infobox,
                field_summary=field_summary,
                field_5=field_5,
                subject_idx_cn=subject_idx_cn,
                subject_airtime=subject_airtime,
                subject_nsfw=subject_nsfw,
                subject_ban=subject_ban,
            )
        )

        db_session.execute(
            sa.delete(ChiiSubjectField).where(ChiiSubjectField.field_sid == subject_id)
        )
        db_session.add(
            ChiiSubjectField(
                field_sid=subject_id,
                field_tags=field_tags,
                field_airtime=field_airtime,
                field_year=field_year,
                field_mon=field_mon,
                field_week_day=field_week_day,
                field_date=field_date,
            )
        )
        db_session.commit()

    try:
        yield mock_id
    finally:
        db_session.execute(
            sa.delete(ChiiSubjectField).where(
                ChiiSubjectField.field_sid.in_(mock_subject_id)
            )
        )
        db_session.execute(
            sa.delete(ChiiSubject).where(ChiiSubject.subject_id.in_(mock_subject_id))
        )
        db_session.commit()
