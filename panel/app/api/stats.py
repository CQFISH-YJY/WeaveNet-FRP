"""统计 API：流量统计、在线概览。"""
from __future__ import annotations

from datetime import date, timedelta

from fastapi import APIRouter, Depends
from sqlalchemy import func
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import success
from ..core.redis_client import redis_client
from ..models import TrafficStat, Tunnel, User

router = APIRouter(prefix="/api/stats", tags=["统计"])


@router.get("/traffic")
def traffic_stats(
    days: int = 7,
    tunnel_id: int | None = None,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """流量统计（按日聚合）。"""
    if days > 31:
        days = 31
    start_date = (date.today() - timedelta(days=days - 1)).strftime("%Y-%m-%d")
    q = (
        db.query(
            TrafficStat.date,
            func.sum(TrafficStat.in_bytes).label("in"),
            func.sum(TrafficStat.out_bytes).label("out"),
        )
        .join(Tunnel, TrafficStat.tunnel_id == Tunnel.id)
        .filter(Tunnel.user_id == user.id, TrafficStat.date >= start_date)
    )
    if tunnel_id:
        q = q.filter(Tunnel.id == tunnel_id)
    rows = q.group_by(TrafficStat.date).all()

    result = []
    for d in range(days):
        day = (date.today() - timedelta(days=days - 1 - d)).strftime("%Y-%m-%d")
        matched = next((r for r in rows if r.date == day), None)
        result.append(
            {
                "date": day,
                "in_bytes": int(getattr(matched, "in", 0) or 0) if matched else 0,
                "out_bytes": int(getattr(matched, "out", 0) or 0) if matched else 0,
            }
        )
    return success(result)


@router.get("/overview")
def overview(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """在线概览。"""
    tunnels = db.query(Tunnel).filter(Tunnel.user_id == user.id).all()
    running = 0
    total_in = 0
    total_out = 0
    total_conn = 0
    for t in tunnels:
        runtime = redis_client.get_tunnel_runtime(t.id) or {}
        if runtime.get("online") or t.status == "running":
            running += 1
        total_in += int(runtime.get("in", 0))
        total_out += int(runtime.get("out", 0))
        total_conn += int(runtime.get("connections", 0))

    return success(
        {
            "tunnel_total": len(tunnels),
            "tunnel_running": running,
            "today_in": total_in,
            "today_out": total_out,
            "connections": total_conn,
            "points": user.points,
        }
    )
