from databases import Database, DatabaseURL
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from pol import config

database = Database(config.MYSQL_URI, force_rollback=config.TESTING)

engine = create_engine(
    str(DatabaseURL(config.MYSQL_URI).replace(dialect="mysql+pymysql"))
)
Session = sessionmaker(bind=engine)
