"""应用日志配置。"""
from __future__ import annotations

import logging
import sys

from .config import get_settings

settings = get_settings()

LOGGING_CONFIG: dict = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "default": {
            "format": "%(asctime)s [%(levelname)s] %(name)s: %(message)s",
            "datefmt": "%Y-%m-%d %H:%M:%S",
        }
    },
    "handlers": {
        "console": {
            "class": "logging.StreamHandler",
            "stream": sys.stdout,
            "formatter": "default",
        }
    },
    "root": {"level": "DEBUG" if settings.debug else "INFO", "handlers": ["console"]},
    "loggers": {
        "weavenet": {
            "level": "DEBUG" if settings.debug else "INFO",
            "handlers": ["console"],
            "propagate": False,
        },
        "uvicorn": {"handlers": ["console"], "level": "INFO", "propagate": False},
        "uvicorn.error": {"handlers": ["console"], "level": "INFO", "propagate": False},
        "sqlalchemy.engine": {"handlers": ["console"], "level": "WARNING", "propagate": False},
    },
}


def get_logger(name: str = "weavenet") -> logging.Logger:
    """获取命名日志器。"""
    return logging.getLogger(name)
