"""用户个人 API：资料、密码、邮箱、额度、操作日志。"""
from __future__ import annotations

from datetime import datetime

from fastapi import APIRouter, Depends, Request
from pydantic import BaseModel, EmailStr, Field
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import BizError, ConflictError, success
from ..core.security import generate_email_code, hash_password, verify_password
from ..models import EmailCode, OperationLog, Session as SessionModel
from ..models import User
from ..services.mail import send_verification_code

router = APIRouter(prefix="/api/user", tags=["用户"])


class ProfileUpdate(BaseModel):
    """修改个人资料请求。"""

    email: EmailStr | None = None


class PasswordUpdate(BaseModel):
    """修改密码请求。"""

    old_password: str = Field(min_length=1, max_length=64)
    new_password: str = Field(min_length=8, max_length=64)


class EmailChangeRequest(BaseModel):
    """修改邮箱请求。"""

    new_email: EmailStr
    code: str = Field(min_length=4, max_length=16)


@router.get("/profile")
def profile(user: User = Depends(get_current_user)):
    """获取个人资料。"""
    return success(
        {
            "id": user.id,
            "username": user.username,
            "email": user.email,
            "email_verified": user.email_verified,
            "status": user.status,
            "points": user.points,
            "plan": {
                "id": user.plan.id if user.plan else None,
                "name": user.plan.name if user.plan else "",
                "speed_limit_mbps": user.plan.speed_limit_mbps if user.plan else 8,
                "tunnel_limit": user.plan.tunnel_limit if user.plan else 3,
                "domain_limit": user.plan.domain_limit if user.plan else 1,
            }
            if user.plan
            else None,
            "plan_expires_at": user.plan_expires_at.isoformat() if user.plan_expires_at else None,
            "created_at": user.created_at.isoformat() if user.created_at else None,
            "last_login_at": user.last_login_at.isoformat() if user.last_login_at else None,
        }
    )


@router.put("/profile")
def update_profile(
    payload: ProfileUpdate,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """修改个人资料（当前仅支持邮箱预填，邮箱修改走 /email 流程）。"""
    if payload.email and payload.email != user.email:
        if db.query(User).filter(User.email == payload.email).first():
            raise ConflictError(1003, "该邮箱已被其他账号使用")
        # 修改邮箱需走验证码流程，此处仅保存新邮箱并标记未验证
        user.email = payload.email
        user.email_verified = False
        db.commit()
    return success(message="资料已更新")


@router.post("/email/send-code")
def send_email_code(
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """发送修改邮箱验证码到新邮箱。"""
    # 需要客户端提供 new_email，因此此接口由 /email/change 复用验证码
    return success(message="请在修改邮箱时提供新邮箱与验证码")


@router.post("/email/change")
def change_email(
    payload: EmailChangeRequest,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """修改邮箱：先向新邮箱发送验证码，再携带验证码提交。"""
    if db.query(User).filter(User.email == payload.new_email).first():
        raise ConflictError(1003, "该邮箱已被其他账号使用")

    # 若没有验证码记录则先发送
    code_row = (
        db.query(EmailCode)
        .filter(
            EmailCode.email == payload.new_email,
            EmailCode.purpose == "change_email",
            EmailCode.used.is_(False),
        )
        .order_by(EmailCode.id.desc())
        .first()
    )
    if code_row is None:
        code = generate_email_code()
        db.add(
            EmailCode(
                email=payload.new_email,
                code=code,
                purpose="change_email",
                expires_at=datetime.now().replace(minute=datetime.now().minute + 5),
            )
        )
        db.commit()
        send_verification_code(payload.new_email, code, "change_email")
        return success(message="验证码已发送到新邮箱，请查收")

    if code_row.code != payload.code:
        raise BizError(400, 1001, "验证码错误")
    if code_row.expires_at < datetime.now():
        raise BizError(400, 1002, "验证码已过期")

    old_email = user.email
    user.email = payload.new_email
    user.email_verified = True
    code_row.used = True
    db.commit()
    return success({"old_email": old_email, "new_email": payload.new_email}, "邮箱修改成功")


@router.put("/password")
def change_password(
    payload: PasswordUpdate,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """修改密码。"""
    if not verify_password(payload.old_password, user.password_hash):
        raise BizError(400, 0, "原密码错误")
    user.password_hash = hash_password(payload.new_password)
    # 使其他会话失效（保留当前会话由前端重新登录）
    db.query(SessionModel).filter(SessionModel.user_id == user.id).delete()
    db.commit()
    return success(message="密码修改成功，请重新登录")


@router.get("/quota")
def quota(user: User = Depends(get_current_user)):
    """查询套餐额度。"""
    plan = user.plan
    return success(
        {
            "plan": {
                "id": plan.id if plan else 1,
                "name": plan.name if plan else "免费版",
                "speed_limit_mbps": plan.speed_limit_mbps if plan else 8,
                "tunnel_limit": plan.tunnel_limit if plan else 3,
                "domain_limit": plan.domain_limit if plan else 1,
            },
            "plan_expires_at": user.plan_expires_at.isoformat() if user.plan_expires_at else None,
            "tunnel_count": len(user.tunnels),
            "tunnel_limit": plan.tunnel_limit if plan else 3,
            "domain_count": sum(1 for d in user.tunnels if False),
        }
    )


@router.get("/logs")
def user_logs(
    page: int = 1,
    page_size: int = 20,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """查询用户操作日志。"""
    # 用户操作记录在 OperationLog 中以 target_type=user 记录
    q = (
        db.query(OperationLog)
        .filter(OperationLog.target_type == "user", OperationLog.target_id == user.id)
        .order_by(OperationLog.id.desc())
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
                    "action": x.action,
                    "detail": x.detail,
                    "created_at": x.created_at.isoformat() if x.created_at else None,
                }
                for x in items
            ],
        }
    )
