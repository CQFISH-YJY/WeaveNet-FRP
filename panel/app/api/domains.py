"""免费域名 API：申请/释放。"""
from __future__ import annotations

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.config import get_settings
from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import BizError, ConflictError, NotFoundError, success
from ..models import Domain, Tunnel, User
from ..schemas.schemas import DomainCreate
from ..services.frp_config import check_domain_limit, check_subdomain_available

router = APIRouter(prefix="/api/domains", tags=["免费域名"])
settings = get_settings()


@router.get("")
def list_domains(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """免费域名列表。"""
    domains = (
        db.query(Domain)
        .filter(Domain.user_id == user.id)
        .order_by(Domain.id.desc())
        .all()
    )
    return success(
        [
            {
                "id": d.id,
                "user_id": d.user_id,
                "tunnel_id": d.tunnel_id,
                "subdomain": d.subdomain,
                "full_domain": d.full_domain,
                "status": d.status,
                "created_at": d.created_at.isoformat() if d.created_at else None,
            }
            for d in domains
        ]
    )


@router.post("")
def create_domain(
    payload: DomainCreate,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """申请免费域名。"""
    db.refresh(user)
    plan = user.plan
    if plan is None:
        raise BizError(400, 2001, "套餐配置异常，请联系管理员")
    check_domain_limit(user, plan)

    if not check_subdomain_available(db, payload.subdomain):
        raise ConflictError(3002, "子域名已被占用，请更换")

    domain = Domain(
        user_id=user.id,
        tunnel_id=payload.tunnel_id,
        subdomain=payload.subdomain,
        full_domain=f"{payload.subdomain}.{settings.domain_suffix}",
        status="active",
    )
    # 校验域名路由与隧道绑定（若指定隧道，隧道须属于该用户且为 http/https 类型）
    if payload.tunnel_id:
        tunnel = (
            db.query(Tunnel)
            .filter(Tunnel.id == payload.tunnel_id, Tunnel.user_id == user.id)
            .first()
        )
        if tunnel is None:
            raise NotFoundError(message="隧道不存在")
        if tunnel.type not in ("http", "https"):
            raise BizError(400, 0, "仅 HTTP/HTTPS 隧道可绑定免费域名")
        tunnel.subdomain = payload.subdomain
    db.add(domain)
    db.commit()
    db.refresh(domain)
    return success(
        {
            "id": domain.id,
            "subdomain": domain.subdomain,
            "full_domain": domain.full_domain,
        },
        "免费域名申请成功",
        201,
    )


@router.delete("/{domain_id}", status_code=204)
def release_domain(
    domain_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """释放免费域名。"""
    domain = (
        db.query(Domain)
        .filter(Domain.id == domain_id, Domain.user_id == user.id)
        .first()
    )
    if domain is None:
        raise NotFoundError(message="域名不存在")
    domain.status = "released"
    # 解除隧道子域名绑定
    if domain.tunnel_id:
        tunnel = db.get(Tunnel, domain.tunnel_id)
        if tunnel and tunnel.subdomain == domain.subdomain:
            tunnel.subdomain = None
    db.commit()
