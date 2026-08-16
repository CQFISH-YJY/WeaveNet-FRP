"""SQLite 数据库封装。

启用 WAL 模式与 busy_timeout，写操作自动重试缓解锁冲突。
"""
from __future__ import annotations

from collections.abc import Generator

from sqlalchemy import create_engine, event
from sqlalchemy.orm import DeclarativeBase, Session, sessionmaker

from .config import get_settings

settings = get_settings()

# connect_args: SQLite 专用参数；其他数据库忽略
_kwargs: dict = {"check_same_thread": False}
if settings.database_url.startswith("sqlite"):
    _kwargs["timeout"] = 30

engine = create_engine(
    settings.database_url,
    connect_args=_kwargs,
    pool_pre_ping=True,
)


@event.listens_for(engine, "connect")
def _set_sqlite_pragma(dbapi_connection, connection_record) -> None:  # noqa: ANN001
    """SQLite 连接时启用 WAL 与 busy_timeout。"""
    cursor = dbapi_connection.cursor()
    cursor.execute("PRAGMA journal_mode=WAL")
    cursor.execute("PRAGMA busy_timeout=30000")
    cursor.execute("PRAGMA synchronous=NORMAL")
    cursor.close()


SessionLocal = sessionmaker(bind=engine, autoflush=False, expire_on_commit=False)


class Base(DeclarativeBase):
    """ORM 基类。"""


def get_db() -> Generator[Session, None, None]:
    """FastAPI 依赖：提供数据库会话并保证关闭。"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def retry_on_locked(max_retries: int = 3):
    """SQLite 锁冲突时自动重试的装饰器。

    SQLite 处于 WAL 模式后绝大多数并发读写可正常工作，但极端情况下
    仍可能出现 database is locked，此处提供兜底重试。
    """

    def decorator(func):
        def wrapper(*args, **kwargs):
            import sqlite3

            from sqlalchemy.exc import OperationalError

            for attempt in range(max_retries):
                try:
                    return func(*args, **kwargs)
                except OperationalError as exc:
                    if "locked" not in str(exc) or attempt == max_retries - 1:
                        raise
                    import time

                    time.sleep(0.2 * (attempt + 1))
                except sqlite3.OperationalError as exc:
                    if "locked" not in str(exc) or attempt == max_retries - 1:
                        raise
                    import time

                    time.sleep(0.2 * (attempt + 1))

        return wrapper

    return decorator
