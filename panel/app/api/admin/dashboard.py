"""管理后台：全局隧道管理、套餐配置、公告、系统配置、日志、统计看板。"""
from __future__ import annotations

from datetime import date, timedelta

from fastapi import APIRouter, Depends, Request
from sqlalchemy import func
from sqlalchemy.orm import Session

from ...core.database import get_db
from ...core.deps import get_current_admin
from ...core.errors import ConflictError, NotFoundError, success
from ...core.redis_client import redis_client
from ...models import (
    Announcement,
    Node,
    OperationLog,
    Plan,
    PointsLog,
    SystemConfig,
    TrafficStat,
    Tunnel,
    User,
)
from ...schemas.schemas import (
    AnnouncementCreate,
    AnnouncementUpdate,
    ConfigUpdate,
    PlanUpdate,
)

router = APIRouter(prefix="/api/admin", tags=["管理"])


def _log(db: Session, request: Request, admin: User, action: str, target_type: str, target_id: int | None, detail: str) -> None:  # noqa: ANN001
    db.add(
        OperationLog(
            admin_id=admin.id,
            admin_name=admin.username,
            action=action,
            target_type=target_type,
            target_id=target_id,
            detail=detail,
            ip=request.client.host if request.client else "",
        )
    )


# ---------- 全局隧道 ----------

@router.get("/tunnels")
def admin_tunnels(
    keyword: str | None = None,
    node_id: int | None = None,
    status: str | None = None,
    page: int = 1,
    page_size: int = 20,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """全局隧道列表。"""
    q = db.query(Tunnel).join(User).join(Node)
    if keyword:
        q = q.filter(
            (Tunnel.name.like(f"%{keyword}%"))
            | (Tunnel.local_ip.like(f"%{keyword}%"))
            | (User.username.like(f"%{keyword}%"))
        )
    if node_id:
        q = q.filter(Tunnel.node_id == node_id)
    if status:
        q = q.filter(Tunnel.status == status)
    total = q.count()
    items = q.order_by(Tunnel.id.desc()).offset((page - 1) * page_size).limit(page_size).all()
    return success(
        {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [
                {
                    "id": t.id,
                    "name": t.name,
                    "type": t.type,
                    "username": t.user.username,
                    "node_name": t.node.name if t.node else "",
                    "local_ip": t.local_ip,
                    "local_port": t.local_port,
                    "remote_port": t.remote_port,
                    "subdomain": t.subdomain,
                    "status": t.status,
                    "created_at": t.created_at.isoformat() if t.created_at else None,
                }
                for t in items
            ],
        }
    )


@router.post("/tunnels/{tunnel_id}/offline")
def force_offline_tunnel(
    tunnel_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """强制下线隧道。"""
    tunnel = db.get(Tunnel, tunnel_id)
    if tunnel is None:
        raise NotFoundError(message="隧道不存在")
    tunnel.status = "stopped"
    tunnel.status_detail = "管理员强制下线"
    redis_client.set_json(f"tunnel:want:{tunnel.id}", 3600, {"action": "stop", "force": True})
    redis_client.delete(f"tunnel:runtime:{tunnel.id}")
    _log(db, request, admin, "force_offline_tunnel", "tunnel", tunnel.id, f"强制下线隧道 {tunnel.name}")
    db.commit()
    return success(message="隧道已强制下线")


# ---------- 套餐配置 ----------

@router.get("/plans")
def list_plans(admin: User = Depends(get_current_admin), db: Session = Depends(get_db)):
    """套餐列表。"""
    plans = db.query(Plan).order_by(Plan.sort).all()
    return success(
        [
            {
                "id": p.id,
                "name": p.name,
                "speed_limit_mbps": p.speed_limit_mbps,
                "tunnel_limit": p.tunnel_limit,
                "domain_limit": p.domain_limit,
                "sort": p.sort,
            }
            for p in plans
        ]
    )


@router.put("/plans/{plan_id}")
def update_plan(
    plan_id: int,
    payload: PlanUpdate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """修改套餐档位。"""
    plan = db.get(Plan, plan_id)
    if plan is None:
        raise NotFoundError(message="套餐不存在")
    if plan.is_default:
        raise ConflictError(0, "免费版为默认套餐，不可修改额度")
    for field in ("name", "speed_limit_mbps", "tunnel_limit", "domain_limit", "sort"):
        value = getattr(payload, field, None)
        if value is not None:
            setattr(plan, field, value)
    _log(db, request, admin, "update_plan", "plan", plan.id, f"修改套餐 {plan.name}")
    db.commit()
    db.refresh(plan)
    return success(
        {
            "id": plan.id,
            "name": plan.name,
            "speed_limit_mbps": plan.speed_limit_mbps,
            "tunnel_limit": plan.tunnel_limit,
            "domain_limit": plan.domain_limit,
            "sort": plan.sort,
        },
        "套餐已更新",
    )


# ---------- 公告 ----------

@router.post("/announcements")
def create_announcement(
    payload: AnnouncementCreate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """发布公告。"""
    announcement = Announcement(
        title=payload.title,
        content=payload.content,
        author=payload.author,
        status="active",
    )
    db.add(announcement)
    _log(db, request, admin, "create_announcement", "announcement", None, f"发布公告《{payload.title}》")
    db.commit()
    db.refresh(announcement)
    return success(
        {
            "id": announcement.id,
            "title": announcement.title,
            "author": announcement.author,
            "created_at": announcement.created_at.isoformat() if announcement.created_at else None,
        },
        "公告发布成功",
        201,
    )


@router.get("/announcements")
def admin_announcements(
    page: int = 1,
    page_size: int = 20,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """公告列表（含下线）。"""
    q = db.query(Announcement).order_by(Announcement.id.desc())
    total = q.count()
    items = q.offset((page - 1) * page_size).limit(page_size).all()
    return success(
        {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [
                {
                    "id": a.id,
                    "title": a.title,
                    "content": a.content,
                    "author": a.author,
                    "status": a.status,
                    "created_at": a.created_at.isoformat() if a.created_at else None,
                }
                for a in items
            ],
        }
    )


@router.put("/announcements/{announcement_id}")
def update_announcement(
    announcement_id: int,
    payload: AnnouncementUpdate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """修改公告。"""
    announcement = db.get(Announcement, announcement_id)
    if announcement is None:
        raise NotFoundError(message="公告不存在")
    for field in ("title", "content", "author"):
        value = getattr(payload, field, None)
        if value is not None:
            setattr(announcement, field, value)
    _log(db, request, admin, "update_announcement", "announcement", announcement.id, f"修改公告《{announcement.title}》")
    db.commit()
    return success(message="公告已更新")


@router.post("/announcements/{announcement_id}/offline")
def offline_announcement(
    announcement_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """公告下线。"""
    announcement = db.get(Announcement, announcement_id)
    if announcement is None:
        raise NotFoundError(message="公告不存在")
    announcement.status = "offline"
    _log(db, request, admin, "offline_announcement", "announcement", announcement.id, f"下线公告《{announcement.title}》")
    db.commit()
    return success(message="公告已下线")


# ---------- 系统配置 ----------

@router.get("/config")
def get_config(admin: User = Depends(get_current_admin), db: Session = Depends(get_db)):
    """获取系统配置。"""
    configs = {c.key: c.value for c in db.query(SystemConfig).all()}
    return success(configs)


@router.put("/config")
def update_config(
    payload: ConfigUpdate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """修改系统配置。"""
    row = db.get(SystemConfig, payload.key)
    if row is None:
        row = SystemConfig(key=payload.key, value=payload.value)
        db.add(row)
    else:
        row.value = payload.value
    _log(db, request, admin, "update_config", "config", None, f"修改配置 {payload.key}={payload.value}")
    db.commit()
    return success(message="配置已保存")


# ---------- 日志 ----------

@router.get("/logs/operation")
def operation_logs(
    page: int = 1,
    page_size: int = 50,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """操作日志。"""
    q = db.query(OperationLog).order_by(OperationLog.id.desc())
    total = q.count()
    items = q.offset((page - 1) * page_size).limit(page_size).all()
    return success(
        {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [
                {
                    "id": x.id,
                    "admin_name": x.admin_name,
                    "action": x.action,
                    "target_type": x.target_type,
                    "target_id": x.target_id,
                    "detail": x.detail,
                    "ip": x.ip,
                    "created_at": x.created_at.isoformat() if x.created_at else None,
                }
                for x in items
            ],
        }
    )


@router.get("/logs/runtime")
def runtime_logs(
    lines: int = 200,
    admin: User = Depends(get_current_admin),
):
    """运行日志（从本地日志文件读取尾部）。"""
    import os

    from ...core.logger import settings as log_settings  # noqa: F401

    log_dir = log_settings.BASE_DIR / "logs"
    log_file = log_dir / "weavenet.log"
    if not log_file.exists():
        return success({"lines": [], "file": str(log_file)})
    try:
        with open(log_file, "r", encoding="utf-8", errors="ignore") as f:
            content = f.readlines()
        tail = content[-max(0, lines):]
        return success({"lines": [x.rstrip("\n") for x in tail], "file": str(log_file)})
    except OSError:
        return success({"lines": [], "file": str(log_file)})


# ---------- 统计看板 ----------

@router.get("/dashboard")
def dashboard(
    days: int = 7,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """统计看板。"""
    today = date.today()
    start = today - timedelta(days=days - 1)

    total_users = db.query(User).count()
    new_users_week = (
        db.query(User).filter(User.created_at >= start).count()
    )
    online_nodes = db.query(Node).filter(Node.status == "online").count()
    total_nodes = db.query(Node).count()
    running_tunnels = db.query(Tunnel).filter(Tunnel.status == "running").count()
    total_tunnels = db.query(Tunnel).count()

    traffic_rows = (
        db.query(
            TrafficStat.date,
            func.sum(TrafficStat.in_bytes).label("in"),
            func.sum(TrafficStat.out_bytes).label("out"),
        )
        .filter(TrafficStat.date >= start.strftime("%Y-%m-%d"))
        .group_by(TrafficStat.date)
        .all()
    )
    traffic_map = {r.date: r for r in traffic_rows}
    traffic_series = []
    for d in range(days):
        day = (today - timedelta(days=days - 1 - d)).strftime("%Y-%m-%d")
        r = traffic_map.get(day)
        traffic_series.append(
            {
                "date": day,
                "in_bytes": int(getattr(r, "in", 0) or 0) if r else 0,
                "out_bytes": int(getattr(r, "out", 0) or 0) if r else 0,
            }
        )

    points_issued = (
        db.query(func.sum(PointsLog.change))
        .filter(PointsLog.change > 0)
        .scalar()
        or 0
    )

    return success(
        {
            "summary": {
                "total_users": total_users,
                "new_users_week": new_users_week,
                "online_nodes": online_nodes,
                "total_nodes": total_nodes,
                "running_tunnels": running_tunnels,
                "total_tunnels": total_tunnels,
                "points_issued": int(points_issued),
            },
            "traffic_series": traffic_series,
        }
    )
