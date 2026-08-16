"""管理后台：用户管理 API。"""
from __future__ import annotations

from datetime import datetime, timedelta

from fastapi import APIRouter, Depends, Request
from sqlalchemy.orm import Session

from ...core.database import get_db
from ...core.deps import get_current_admin
from ...core.errors import NotFoundError, success
from ...core.security import generate_token, hash_password
from ...models import OperationLog, Plan, User, UserPlanLog
from ...schemas.schemas import AdminUserBan, AdminUserUpdate

router = APIRouter(prefix="/api/admin/users", tags=["管理-用户"])


def _log_operation(db: Session, request: Request, admin: User, action: str, target_type: str, target_id: int | None, detail: str) -> None:  # noqa: ANN001
    """写入操作日志。"""
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


@router.get("")
def list_users(
    keyword: str | None = None,
    status: str | None = None,
    page: int = 1,
    page_size: int = 20,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """用户列表。"""
    q = db.query(User)
    if keyword:
        q = q.filter(
            (User.username.like(f"%{keyword}%")) | (User.email.like(f"%{keyword}%"))
        )
    if status:
        q = q.filter(User.status == status)
    total = q.count()
    items = q.order_by(User.id.desc()).offset((page - 1) * page_size).limit(page_size).all()
    return success(
        {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [
                {
                    "id": u.id,
                    "username": u.username,
                    "email": u.email,
                    "email_verified": u.email_verified,
                    "status": u.status,
                    "points": u.points,
                    "plan_name": u.plan.name if u.plan else "",
                    "plan_expires_at": u.plan_expires_at.isoformat() if u.plan_expires_at else None,
                    "created_at": u.created_at.isoformat() if u.created_at else None,
                    "last_login_at": u.last_login_at.isoformat() if u.last_login_at else None,
                }
                for u in items
            ],
        }
    )


@router.post("/{user_id}/ban")
def ban_user(
    user_id: int,
    payload: AdminUserBan,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """封禁用户。"""
    user = db.get(User, user_id)
    if user is None:
        raise NotFoundError(message="用户不存在")
    user.status = "banned"
    detail = payload.reason or "管理员封禁"
    _log_operation(db, request, admin, "ban_user", "user", user.id, detail)
    db.commit()
    return success(message=f"已封禁用户 {user.username}")


@router.post("/{user_id}/unban")
def unban_user(
    user_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """解除封禁。"""
    user = db.get(User, user_id)
    if user is None:
        raise NotFoundError(message="用户不存在")
    user.status = "active"
    _log_operation(db, request, admin, "unban_user", "user", user.id, "")
    db.commit()
    return success(message=f"已解除封禁 {user.username}")


@router.post("/{user_id}/reset-password")
def reset_password(
    user_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """重置密码（生成随机临时密码）。"""
    user = db.get(User, user_id)
    if user is None:
        raise NotFoundError(message="用户不存在")
    temp_password = generate_token(12)
    user.password_hash = hash_password(temp_password)
    _log_operation(db, request, admin, "reset_password", "user", user.id, "重置用户密码")
    db.commit()
    return success({"temp_password": temp_password}, "密码已重置，请将临时密码告知用户")


@router.put("/{user_id}/plan")
def update_user_plan(
    user_id: int,
    payload: AdminUserUpdate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """调整用户套餐或积分。"""
    user = db.get(User, user_id)
    if user is None:
        raise NotFoundError(message="用户不存在")
    if payload.plan_id is not None:
        plan = db.get(Plan, payload.plan_id)
        if plan is None:
            raise NotFoundError(message="套餐不存在")
        user.plan_id = plan.id
        # 手动调整套餐：按 30 天计
        user.plan_expires_at = datetime.now() + timedelta(days=30)
        db.add(
            UserPlanLog(
                user_id=user.id,
                plan_id=plan.id,
                plan_name=plan.name,
                reason="管理员调整",
            )
        )
        _log_operation(db, request, admin, "update_user_plan", "user", user.id, f"调整套餐为 {plan.name}")
    if payload.points is not None:
        user.points = payload.points
        _log_operation(db, request, admin, "update_user_points", "user", user.id, f"调整积分为 {payload.points}")
    db.commit()
    return success(message="用户信息已更新")
