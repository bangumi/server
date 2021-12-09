from databases import Database

from pol import config

database = Database(config.MYSQL_URI, force_rollback=config.TESTING)
