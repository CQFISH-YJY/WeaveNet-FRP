"""积分 API：积分流水、兑换会员。"""
from __future__ import annotations

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.config import get_settings
from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import success
from ..models import PointsLog, SystemConfig, User
from ..services.points import exchange_membership

router = APIRouter(prefix="/api/points", tags=["积分"])
settings = get_settings()


@router.get("/logs")
def points_logs(
    page: int = 1,
    page_size: int = 20,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """积分流水。"""
    q = (
        db.query(PointsLog)
        .filter(PointsLog.user_id == user.id)
        .order_by(PointsLog.id.desc())
    )
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
                    "change": x.change,
                    "balance": x.balance,
                    "reason": x.reason,
                    "created_at": x.created_at.isoformat() if x.created_at else None,
                }
                for x in items
            ],
        }
    )


@router.post("/exchange")
def exchange(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """积分兑换会员。"""
    result = exchange_membership(db, user)
    return success(result, "兑换成功")


@router.get("/rules")
def points_rules(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """查询当前积分规则。"""
    config = {c.key: c.value for c in db.query(SystemConfig).all()}
    return success(
        {
            "signin_points": int(config.get("signin_points", settings.signin_points)),
            "signin_streak_bonus": int(config.get("signin_streak_bonus", settings.signin_streak_bonus)),
            "signin_streak_days": int(config.get("signin_streak_days", settings.signin_streak_days)),
            "exchange_points": int(config.get("exchange_points", settings.exchange_points)),
            "exchange_plan_days": int(config.get("exchange_plan_days", settings.exchange_plan_days)),
            "exchange_plan_name": config.get("exchange_plan_name", "普通会员"),
        }
    )
