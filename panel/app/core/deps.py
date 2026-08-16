"""FastAPI 依赖：当前用户、角色校验、节点 Agent、限流。"""
from __future__ import annotations

from datetime import datetime

from fastapi import Depends, Request
from sqlalchemy.orm import Session

from ..models import Session as SessionModel
from ..models import User
from .config import get_settings
from .database import get_db
from .errors import AuthError, ForbiddenError, RateLimitedError
from .redis_client import redis_client

settings = get_settings()

ADMIN_ROLE = "admin"
USER_ROLE = "user"


def _extract_token(request: Request) -> str:
    """从请求头提取 Bearer Token。"""
    auth = request.headers.get("Authorization", "")
    if auth.startswith("Bearer "):
        return auth[len("Bearer ") :].strip()
    return ""


def _session_valid(db: Session, token: str, user_id: int) -> bool:
    """校验会话（Redis 优先，SQLite 兜底双写）。"""
    # 通过 Redis 判断会话仍在有效期
    if redis_client.get_session_user(token) is not None:
        return True
    # Redis 降级或丢失时查询数据库
    row = (
        db.query(SessionModel)
        .filter(
            SessionModel.token == token,
            SessionModel.user_id == user_id,
        )
        .first()
    )
    if row and row.expires_at > datetime.now():
        # 回写 Redis
        redis_client.set_session(token, user_id)
        return True
    return False


def get_current_user(
    request: Request,
    db: Session = Depends(get_db),
) -> User:
    """获取当前登录用户，失败抛 401。"""
    token = _extract_token(request)
    if not token:
        raise AuthError(message="未登录或登录已过期")
    session_data = redis_client.get_json(f"session:{token}")
    if session_data is not None:
        user_id = int(session_data.get("user_id", 0))
        row = db.get(User, user_id)
        if row is None or not _session_valid(db, token, user_id):
            raise AuthError(message="未登录或登录已过期")
        return row
    row = (
        db.query(SessionModel)
        .filter(SessionModel.token == token)
        .first()
    )
    if row is None:
        raise AuthError(message="未登录或登录已过期")
    user = db.get(User, row.user_id)
    if user is None or row.expires_at <= datetime.now():
        raise AuthError(message="未登录或登录已过期")
    redis_client.set_session(token, user.id)
    if user.status == "banned":
        raise ForbiddenError(1006, "账号已被封禁，请联系管理员")
    return user


def get_current_admin(
    request: Request,
    db: Session = Depends(get_db),
) -> User:
    """获取当前管理员，失败抛 403。"""
    user = get_current_user(request, db)
    if user.username != settings.admin_username and not getattr(user, "is_admin", False):
        # 管理员由初始化账号唯一标识；用户名匹配即视为 admin
        raise ForbiddenError(message="需要管理员权限")
    return user


def get_agent_node(
    request: Request,
    db: Session = Depends(get_db),
):
    """校验 Agent Token，返回对应节点。"""
    token = _extract_token(request)
    if not token or not token.startswith("agent_"):
        raise AuthError(message="Agent 鉴权失败")
    from ..models import Node

    node = db.query(Node).filter(Node.agent_token == token).first()
    if node is None:
        raise AuthError(message="Agent 鉴权失败")
    return node


def require_rate_limit(scope: str, limit: int = 10, window: int = 60):
    """限流依赖工厂：scope 为限流维度前缀。"""

    def dependency(request: Request) -> None:
        if not settings.rate_limit_enabled:
            return
        ip = request.client.host if request.client else "unknown"
        bucket = f"{scope}:{ip}"
        if redis_client.rate_limit_hit(bucket, limit, window):
            raise RateLimitedError(message="请求过于频繁，请稍后再试")

    return dependency
