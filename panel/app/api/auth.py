"""用户认证 API：注册、登录、邮箱验证、找回密码、登出。"""
from __future__ import annotations

from datetime import datetime, timedelta

from fastapi import APIRouter, Depends, Request
from sqlalchemy import or_
from sqlalchemy.orm import Session

from ..core.config import get_settings
from ..core.database import get_db
from ..core.deps import get_current_user, require_rate_limit
from ..core.errors import BizError, ConflictError, NotFoundError, success
from ..core.security import (
    generate_email_code,
    generate_token,
    hash_password,
    verify_password,
)
from ..models import EmailCode, Session as SessionModel
from ..models import User
from ..schemas.auth import (
    EmailVerifyRequest,
    ForgotPasswordRequest,
    LoginRequest,
    RegisterRequest,
    ResetPasswordRequest,
)
from ..services.mail import send_verification_code
from ..core.redis_client import redis_client

router = APIRouter(prefix="/api/auth", tags=["认证"])
settings = get_settings()

EMAIL_CODE_TTL_MINUTES = 5
EMAIL_CODE_MAX_ATTEMPTS = 5


def _create_session(db: Session, user: User) -> str:
    """创建会话：SQLite + Redis 双写。"""
    token = generate_token(64)
    expires_at = datetime.now() + timedelta(days=settings.session_days)
    db.add(SessionModel(token=token, user_id=user.id, expires_at=expires_at))
    db.commit()
    redis_client.set_session(token, user.id, days=settings.session_days)
    return token


@router.post("/register")
def register(
    payload: RegisterRequest,
    request: Request,
    _rl=Depends(require_rate_limit("register", limit=5, window=60)),
    db: Session = Depends(get_db),
):
    """用户注册。"""
    if db.query(User).filter(User.username == payload.username).first():
        raise ConflictError(1004, "用户名已被注册")
    if db.query(User).filter(User.email == payload.email).first():
        raise ConflictError(1003, "邮箱已被注册")

    user = User(
        username=payload.username,
        email=payload.email,
        password_hash=hash_password(payload.password),
        email_verified=False,
        status="active",
        plan_id=1,
        points=0,
    )
    db.add(user)
    db.commit()
    db.refresh(user)

    # 发送注册验证码
    code = generate_email_code()
    expires_at = datetime.now() + timedelta(minutes=EMAIL_CODE_TTL_MINUTES)
    db.add(
        EmailCode(
            email=user.email,
            code=code,
            purpose="register",
            expires_at=expires_at,
        )
    )
    db.commit()
    send_verification_code(user.email, code, "register")

    return success({"id": user.id, "email": user.email}, "注册成功，请查收验证邮件激活账号", 201)


@router.post("/email-verify")
def email_verify(
    payload: EmailVerifyRequest,
    db: Session = Depends(get_db),
):
    """邮箱验证（激活账号 / 换邮箱 / 找回密码第二步由 reset-password 处理）。"""
    code_row = (
        db.query(EmailCode)
        .filter(
            EmailCode.email == payload.email,
            EmailCode.purpose == payload.purpose,
            EmailCode.used.is_(False),
        )
        .order_by(EmailCode.id.desc())
        .first()
    )
    if code_row is None:
        raise BizError(400, 1001, "验证码错误，请重新输入")
    if code_row.expires_at < datetime.now():
        raise BizError(400, 1002, "验证码已过期，请重新获取")
    if code_row.code != payload.code:
        code_row.attempts += 1
        db.commit()
        if code_row.attempts >= EMAIL_CODE_MAX_ATTEMPTS:
            code_row.used = True
            db.commit()
        raise BizError(400, 1001, "验证码错误，请重新输入")

    user = db.query(User).filter(User.email == payload.email).first()
    if user is None:
        raise NotFoundError(1001, "用户不存在")
    user.email_verified = True
    code_row.used = True
    db.commit()
    return success({"email": user.email}, "邮箱验证成功")


@router.post("/resend-code")
def resend_code(
    payload: EmailVerifyRequest,
    request: Request,
    _rl=Depends(require_rate_limit("resend_code", limit=5, window=300)),
    db: Session = Depends(get_db),
):
    """重新发送验证码（5 分钟一次防刷）。"""
    user = db.query(User).filter(User.email == payload.email).first()
    purpose = payload.purpose
    # 注册场景：用户已存在且已验证则提示
    if purpose == "register":
        if user and user.email_verified:
            raise BizError(400, 1005, "邮箱已验证，请直接登录")
    if purpose == "reset_password" and user is None:
        raise NotFoundError(1001, "该邮箱未注册")

    # 校验最近 60 秒内是否已发送（防刷）
    recent = (
        db.query(EmailCode)
        .filter(
            EmailCode.email == payload.email,
            EmailCode.purpose == purpose,
            EmailCode.created_at >= datetime.now() - timedelta(seconds=60),
        )
        .first()
    )
    if recent:
        raise BizError(429, 0, "发送过于频繁，请 1 分钟后再试")

    code = generate_email_code()
    expires_at = datetime.now() + timedelta(minutes=EMAIL_CODE_TTL_MINUTES)
    db.add(
        EmailCode(
            email=payload.email,
            code=code,
            purpose=purpose,
            expires_at=expires_at,
        )
    )
    db.commit()
    send_verification_code(payload.email, code, purpose)
    return success(message="验证码已发送")


@router.post("/login")
def login(
    payload: LoginRequest,
    request: Request,
    _rl=Depends(require_rate_limit("login", limit=10, window=60)),
    db: Session = Depends(get_db),
):
    """用户登录。"""
    user = (
        db.query(User)
        .filter(or_(User.username == payload.username, User.email == payload.username))
        .first()
    )
    if user is None or not verify_password(payload.password, user.password_hash):
        raise BizError(401, 0, "用户名或密码错误")
    if not user.email_verified:
        raise BizError(403, 1005, "邮箱未验证，请先激活账号")
    if user.status == "banned":
        raise BizError(403, 1006, "账号已被封禁，请联系管理员")

    user.last_login_at = datetime.now()
    db.commit()
    token = _create_session(db, user)

    return success(
        {
            "token": token,
            "user": {
                "id": user.id,
                "username": user.username,
                "email": user.email,
                "email_verified": user.email_verified,
                "status": user.status,
                "points": user.points,
                "is_admin": user.username == settings.admin_username,
                "plan_name": user.plan.name if user.plan else "",
                "plan_expires_at": user.plan_expires_at.isoformat() if user.plan_expires_at else None,
            },
        },
        "登录成功",
    )


@router.post("/logout", status_code=204)
def logout(
    request: Request,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """登出。"""
    token = request.headers.get("Authorization", "").replace("Bearer ", "").strip()
    db.query(SessionModel).filter(SessionModel.token == token).delete()
    db.commit()
    redis_client.del_session(token)


@router.post("/forgot-password")
def forgot_password(
    payload: ForgotPasswordRequest,
    request: Request,
    _rl=Depends(require_rate_limit("forgot_password", limit=5, window=300)),
    db: Session = Depends(get_db),
):
    """找回密码第一步：发送验证码。"""
    user = db.query(User).filter(User.email == payload.email).first()
    if user is None:
        # 不暴露邮箱是否存在，统一返回成功
        return success(message="如果该邮箱已注册，验证码已发送")
    if user.status == "banned":
        raise BizError(403, 1006, "账号已被封禁")

    recent = (
        db.query(EmailCode)
        .filter(
            EmailCode.email == payload.email,
            EmailCode.purpose == "reset_password",
            EmailCode.created_at >= datetime.now() - timedelta(seconds=60),
        )
        .first()
    )
    if recent:
        raise BizError(429, 0, "发送过于频繁，请 1 分钟后再试")

    code = generate_email_code()
    expires_at = datetime.now() + timedelta(minutes=EMAIL_CODE_TTL_MINUTES)
    db.add(
        EmailCode(
            email=payload.email,
            code=code,
            purpose="reset_password",
            expires_at=expires_at,
        )
    )
    db.commit()
    send_verification_code(payload.email, code, "reset_password")
    return success(message="如果该邮箱已注册，验证码已发送")


@router.post("/reset-password")
def reset_password(
    payload: ResetPasswordRequest,
    db: Session = Depends(get_db),
):
    """找回密码第二步：验证码校验并重置密码。"""
    code_row = (
        db.query(EmailCode)
        .filter(
            EmailCode.email == payload.email,
            EmailCode.purpose == "reset_password",
            EmailCode.used.is_(False),
        )
        .order_by(EmailCode.id.desc())
        .first()
    )
    if code_row is None:
        raise BizError(400, 1001, "验证码错误，请重新输入")
    if code_row.expires_at < datetime.now():
        raise BizError(400, 1002, "验证码已过期，请重新获取")
    if code_row.code != payload.code:
        raise BizError(400, 1001, "验证码错误，请重新输入")

    user = db.query(User).filter(User.email == payload.email).first()
    if user is None:
        raise NotFoundError(1001, "用户不存在")
    user.password_hash = hash_password(payload.new_password)
    code_row.used = True
    # 重置后使旧会话全部失效
    db.query(SessionModel).filter(SessionModel.user_id == user.id).delete()
    db.commit()
    return success(message="密码重置成功，请使用新密码登录")
