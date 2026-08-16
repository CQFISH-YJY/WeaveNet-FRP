"""客户端 API：Electron 桌面客户端专用。

客户端登录、拉取隧道与节点信息、生成 frpc 配置。
"""
from __future__ import annotations

from fastapi import APIRouter, Depends, Request
from sqlalchemy import or_
from sqlalchemy.orm import Session

from ..core.config import get_settings
from ..core.database import get_db
from ..core.deps import require_rate_limit
from ..core.errors import BizError, NotFoundError, success
from ..core.redis_client import redis_client
from ..core.security import generate_token, verify_password
from ..models import Node, Session as SessionModel
from ..models import Tunnel, User
from ..schemas.auth import LoginRequest
from ..services.frp_config import generate_frpc_config, tunnel_bandwidth

router = APIRouter(prefix="/api/client", tags=["客户端"])
settings = get_settings()


@router.post("/login")
def client_login(
    payload: LoginRequest,
    request: Request,
    _rl=Depends(require_rate_limit("client_login", limit=10, window=60)),
    db: Session = Depends(get_db),
):
    """客户端登录：返回专用 Token 与用户基本信息。"""
    user = (
        db.query(User)
        .filter(or_(User.username == payload.username, User.email == payload.username))
        .first()
    )
    if user is None or not verify_password(payload.password, user.password_hash):
        raise BizError(401, 0, "用户名或密码错误")
    if not user.email_verified:
        raise BizError(403, 1005, "邮箱未验证，请先在网页端激活账号")
    if user.status == "banned":
        raise BizError(403, 1006, "账号已被封禁，请联系管理员")

    token = generate_token(64)
    from datetime import datetime, timedelta

    expires_at = datetime.now() + timedelta(days=settings.session_days)
    db.add(SessionModel(token=token, user_id=user.id, expires_at=expires_at))
    db.commit()
    redis_client.set_session(token, user.id, days=settings.session_days)

    return success(
        {
            "token": token,
            "user": {
                "id": user.id,
                "username": user.username,
                "email": user.email,
                "plan_name": user.plan.name if user.plan else "",
                "plan_expires_at": user.plan_expires_at.isoformat() if user.plan_expires_at else None,
                "speed_limit_mbps": user.plan.speed_limit_mbps if user.plan else 8,
            },
        },
        "登录成功",
    )


@router.get("/tunnels")
def client_tunnels(
    request: Request,
    db: Session = Depends(get_db),
):
    """拉取用户隧道 + 节点信息。"""
    from ..core.deps import get_current_user

    user = get_current_user(request, db)
    tunnels = db.query(Tunnel).filter(Tunnel.user_id == user.id).order_by(Tunnel.id.desc()).all()
    return success(
        {
            "user": {
                "id": user.id,
                "username": user.username,
                "plan_name": user.plan.name if user.plan else "",
                "tunnel_limit": user.plan.tunnel_limit if user.plan else 3,
                "domain_limit": user.plan.domain_limit if user.plan else 1,
                "speed_limit_mbps": user.plan.speed_limit_mbps if user.plan else 8,
            },
            "tunnels": [
                {
                    "id": t.id,
                    "name": t.name,
                    "type": t.type,
                    "node_id": t.node_id,
                    "node_name": t.node.name if t.node else "",
                    "node_address": t.node.address if t.node else "",
                    "node_port": t.node.port if t.node else 7000,
                    "local_ip": t.local_ip,
                    "local_port": t.local_port,
                    "remote_port": t.remote_port,
                    "subdomain": t.subdomain,
                    "custom_domain": t.custom_domain,
                    "kcp": t.kcp,
                    "encryption": t.encryption,
                    "compression": t.compression,
                    "secret_key": t.secret_key,
                    "load_balancers": t.load_balancers,
                    "status": t.status,
                    "bandwidth_limit_kbps": user.plan.speed_limit_mbps * 1000 if user.plan else 8000,
                    "public_address": _public_address(t),
                }
                for t in tunnels
            ],
        }
    )


@router.post("/config")
def client_config(
    payload: dict,
    request: Request,
    db: Session = Depends(get_db),
):
    """生成 frpc.toml 配置。"""
    from ..core.deps import get_current_user

    user = get_current_user(request, db)
    tunnel_id = payload.get("tunnel_id")
    tunnel = (
        db.query(Tunnel)
        .filter(Tunnel.id == tunnel_id, Tunnel.user_id == user.id)
        .first()
    )
    if tunnel is None:
        raise NotFoundError(message="隧道不存在")
    node = db.get(Node, tunnel.node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    config = generate_frpc_config(tunnel, user, node)
    return success({"config": config, "tunnel_id": tunnel.id})


def _public_address(tunnel: Tunnel) -> str:
    node = tunnel.node
    host = node.address if node else settings.panel_base_url
    if tunnel.type in ("http", "https"):
        domain = tunnel.custom_domain or tunnel.subdomain
        if domain:
            scheme = "https" if tunnel.type == "https" else "http"
            return f"{scheme}://{domain}"
        return f"{tunnel.type}://{host}:{tunnel.remote_port or ''}"
    if tunnel.remote_port:
        return f"{host}:{tunnel.remote_port}"
    return host
