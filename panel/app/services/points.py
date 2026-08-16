"""积分与签到业务服务。"""
from __future__ import annotations

from datetime import date, datetime, timedelta

from sqlalchemy.orm import Session

from ..core.config import get_settings
from ..core.errors import BizError
from ..models import PointsLog, SigninLog, User

settings = get_settings()


def add_points(
    db: Session,
    user: User,
    change: int,
    reason: str,
) -> int:
    """增加或扣除积分并记录流水，返回最新余额。"""
    user.points += change
    if user.points < 0:
        user.points -= change
        raise BizError(400, 4002, "积分不足")
    log = PointsLog(
        user_id=user.id,
        change=change,
        balance=user.points,
        reason=reason,
    )
    db.add(log)
    return user.points


def do_signin(db: Session, user: User) -> dict:
    """执行每日签到。

    规则：每日签到得 signin_points 积分；连续签到满 signin_streak_days
    天额外奖励 signin_streak_bonus 积分。
    返回签到结果（积分、连续天数、是否触发连签奖励）。
    """
    today = date.today()
    today_str = today.strftime("%Y-%m-%d")

    existed = (
        db.query(SigninLog)
        .filter(SigninLog.user_id == user.id, SigninLog.signin_date == today_str)
        .first()
    )
    if existed:
        raise BizError(400, 4001, "今天已经签到过了，明天再来吧")

    # 计算连续天数：昨天是否签到
    yesterday = (today - timedelta(days=1)).strftime("%Y-%m-%d")
    last_log = (
        db.query(SigninLog)
        .filter(SigninLog.user_id == user.id)
        .order_by(SigninLog.signin_date.desc())
        .first()
    )
    continuous = 1
    if last_log and last_log.signin_date == yesterday:
        continuous = last_log.continuous_days + 1

    base_points = settings.signin_points
    bonus = 0
    if continuous >= settings.signin_streak_days:
        bonus = settings.signin_streak_bonus

    total = base_points + bonus
    add_points(db, user, total, "每日签到")

    log = SigninLog(
        user_id=user.id,
        signin_date=today_str,
        points=total,
        continuous_days=continuous,
    )
    db.add(log)
    db.commit()

    return {
        "points": total,
        "base_points": base_points,
        "bonus": bonus,
        "continuous_days": continuous,
        "signin_date": today_str,
    }


def can_signin(db: Session, user: User) -> dict:
    """查询今日是否可签到及连续签到信息。"""
    today = date.today().strftime("%Y-%m-%d")
    existed = (
        db.query(SigninLog)
        .filter(SigninLog.user_id == user.id, SigninLog.signin_date == today)
        .first()
    )
    last_log = (
        db.query(SigninLog)
        .filter(SigninLog.user_id == user.id)
        .order_by(SigninLog.signin_date.desc())
        .first()
    )
    continuous = last_log.continuous_days if last_log else 0
    return {
        "today_signed": existed is not None,
        "continuous_days": continuous,
        "next_bonus_in": max(0, settings.signin_streak_days - continuous),
    }


def exchange_membership(db: Session, user: User) -> dict:
    """用积分兑换普通会员。

    规则：exchange_points 积分 = exchange_plan_days 天普通会员。
    """
    from ..models import SystemConfig

    config = {c.key: c.value for c in db.query(SystemConfig).all()}
    cost = int(config.get("exchange_points", settings.exchange_points))
    days = int(config.get("exchange_plan_days", settings.exchange_plan_days))

    if user.points < cost:
        raise BizError(400, 4002, "积分不足，无法兑换会员")

    from ..models import Plan

    normal_plan = (
        db.query(Plan).filter(Plan.name.in_(["普通会员", "普通"])).first()
    )
    if normal_plan is None:
        normal_plan = (
            db.query(Plan).order_by(Plan.sort).offset(1).first()
        )
    if normal_plan is None:
        raise BizError(500, 9001, "套餐配置异常，请联系管理员")

    # 扣除积分
    user.points -= cost
    log = PointsLog(
        user_id=user.id,
        change=-cost,
        balance=user.points,
        reason=f"兑换{normal_plan.name}{days}天",
    )
    db.add(log)

    # 套餐变更：当前套餐到期时间与兑换时间取较晚者
    now = datetime.now()
    base = user.plan_expires_at if user.plan_expires_at and user.plan_expires_at > now else now
    user.plan_id = normal_plan.id
    user.plan_expires_at = base + timedelta(days=days)

    from ..models import UserPlanLog

    db.add(
        UserPlanLog(
            user_id=user.id,
            plan_id=normal_plan.id,
            plan_name=normal_plan.name,
            reason="积分兑换会员",
        )
    )
    db.commit()
    return {
        "plan_name": normal_plan.name,
        "plan_expires_at": user.plan_expires_at.isoformat() if user.plan_expires_at else None,
        "points": user.points,
    }
