import platform

from aiologger.loggers.json import JsonLogger

from pol import config


async def setup_logger():
    logger = JsonLogger(
        name="pol",
        extra={
            "@metadata": {"beat": "py_logging", "version": config.COMMIT_REF},
            "version": config.COMMIT_REF,
            "platform": platform.platform(),
        },
    )

    return logger
