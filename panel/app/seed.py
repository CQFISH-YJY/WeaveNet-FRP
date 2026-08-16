"""数据库种子数据初始化。

首次启动时写入：
- 四档套餐（免费版为默认）
- 管理员账号
- 默认系统配置
- 演示节点
"""
from __future__ import annotations

import logging

from sqlalchemy.orm import Session

from .core.config import get_settings
from .core.security import generate_agent_token, hash_password
from .models import Node, Plan, SystemConfig, User

logger = logging.getLogger("weavenet.seed")
settings = get_settings()


def seed_defaults(db: Session) -> None:
    """写入种子数据（幂等）。"""
    _seed_plans(db)
    _seed_admin(db)
    _seed_configs(db)
    _seed_demo_node(db)
    db.commit()


def _seed_plans(db: Session) -> None:
    if db.query(Plan).count() > 0:
        return
    plans = [
        Plan(id=1, name="免费版", speed_limit_mbps=8, tunnel_limit=3, domain_limit=1, sort=1, is_default=True),
        Plan(id=2, name="普通会员", speed_limit_mbps=16, tunnel_limit=6, domain_limit=4, sort=2),
        Plan(id=3, name="高级会员", speed_limit_mbps=24, tunnel_limit=10, domain_limit=8, sort=3),
        Plan(id=4, name="超级会员", speed_limit_mbps=32, tunnel_limit=14, domain_limit=16, sort=4),
    ]
    db.add_all(plans)
    logger.info("已写入四档套餐")


def _seed_admin(db: Session) -> None:
    if db.query(User).filter(User.username == settings.admin_username).first():
        return
    admin = User(
        username=settings.admin_username,
        email=settings.admin_email,
        password_hash=hash_password(settings.admin_password),
        email_verified=True,
        status="active",
        plan_id=4,
        points=99999,
    )
    db.add(admin)
    logger.info("已创建管理员账号 %s", settings.admin_username)


def _seed_configs(db: Session) -> None:
    defaults = {
        "signin_points": str(settings.signin_points),
        "signin_streak_bonus": str(settings.signin_streak_bonus),
        "signin_streak_days": str(settings.signin_streak_days),
        "exchange_points": str(settings.exchange_points),
        "exchange_plan_days": str(settings.exchange_plan_days),
        "exchange_plan_name": "普通会员",
        "domain_suffix": settings.domain_suffix,
        "smtp_host": settings.smtp_host or "",
        "smtp_port": str(settings.smtp_port),
        "smtp_user": settings.smtp_user or "",
        "smtp_from": settings.smtp_from or "",
        "smtp_use_ssl": "1" if settings.smtp_use_ssl else "0",
        "smtp_use_tls": "1" if settings.smtp_use_tls else "0",
    }
    for key, value in defaults.items():
        if db.get(SystemConfig, key) is None:
            db.add(SystemConfig(key=key, value=value))


def _seed_demo_node(db: Session) -> None:
    node = db.query(Node).first()
    if node is None:
        node = Node(
            name="上海主节点",
            address="127.0.0.1",
            port=7000,
            status="offline",
            speed_limit_mbps=100,
            remark="默认演示节点，配置 Agent Token 后 frps 接入即上线",
        )
        db.add(node)
        logger.info("已创建演示节点")
    if not node.agent_token:
        node.agent_token = generate_agent_token()
        logger.info("已生成演示节点 Agent Token")
