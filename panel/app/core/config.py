"""WeaveNet 织网穿透 全局配置。

配置来源优先级：环境变量 > .env 文件 > 默认值。
"""
from __future__ import annotations

import os
from functools import lru_cache
from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict

BASE_DIR = Path(__file__).resolve().parent.parent.parent


class Settings(BaseSettings):
    """应用配置项。"""

    model_config = SettingsConfigDict(
        env_file=str(BASE_DIR / ".env"),
        env_file_encoding="utf-8",
        extra="ignore",
    )

    # 应用基础
    app_name: str = "WeaveNet 织网穿透"
    debug: bool = False
    secret_key: str = Field(default="weavenet-dev-secret-key-please-change-in-production")

    # 监听与外部地址
    panel_host: str = "0.0.0.0"
    panel_port: int = 8000
    # 面板对外基础地址，用于生成邮件链接与客户端连接
    panel_base_url: str = "http://localhost:8000"

    # 数据库
    database_url: str = Field(default=f"sqlite:///{BASE_DIR / 'weavenet.db'}")

    # Redis
    redis_url: str = "redis://127.0.0.1:6379/0"

    # 会话
    session_days: int = 30

    # 管理员初始化账号
    admin_username: str = "admin"
    admin_password: str = "admin123"
    admin_email: str = "admin@weave.test"

    # SMTP
    smtp_host: str = ""
    smtp_port: int = 465
    smtp_user: str = ""
    smtp_password: str = ""
    smtp_from: str = ""
    smtp_use_tls: bool = True
    smtp_use_ssl: bool = True

    # 免费域名占位
    domain_suffix: str = "weave.test"

    # 套餐与积分默认值
    signin_points: int = 10
    signin_streak_bonus: int = 30
    signin_streak_days: int = 7
    exchange_points: int = 300
    exchange_plan_days: int = 30

    # 节点心跳超时（秒）
    node_heartbeat_timeout: int = 60

    # 安全
    bcrypt_rounds: int = 12
    rate_limit_enabled: bool = True

    @property
    def is_dev(self) -> bool:
        return self.debug or not os.getenv("WEAVE_ENV")


@lru_cache
def get_settings() -> Settings:
    """获取全局单例配置。"""
    return Settings()
