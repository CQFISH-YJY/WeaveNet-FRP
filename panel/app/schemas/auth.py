"""认证相关请求/响应模型。"""
from __future__ import annotations

from pydantic import BaseModel, EmailStr, Field


class RegisterRequest(BaseModel):
    """注册请求。"""

    username: str = Field(min_length=3, max_length=32, pattern=r"^[a-zA-Z0-9_\u4e00-\u9fa5]+$")
    email: EmailStr
    password: str = Field(min_length=8, max_length=64)


class EmailVerifyRequest(BaseModel):
    """邮箱验证请求。"""

    email: EmailStr
    code: str = Field(min_length=4, max_length=16)
    purpose: str = Field(default="register", pattern=r"^(register|reset_password|change_email)$")


class ResendCodeRequest(BaseModel):
    """重新发送验证码请求。"""

    email: EmailStr
    purpose: str = Field(default="register", pattern=r"^(register|reset_password|change_email)$")


class LoginRequest(BaseModel):
    """登录请求。"""

    username: str = Field(min_length=1, max_length=64)
    password: str = Field(min_length=1, max_length=64)


class ForgotPasswordRequest(BaseModel):
    """找回密码请求（第一步：发送验证码）。"""

    email: EmailStr


class ResetPasswordRequest(BaseModel):
    """找回密码请求（第二步：验证并重置）。"""

    email: EmailStr
    code: str = Field(min_length=4, max_length=16)
    new_password: str = Field(min_length=8, max_length=64)


class UserOut(BaseModel):
    """用户信息响应。"""

    id: int
    username: str
    email: str
    email_verified: bool
    status: str
    points: int
    plan_name: str = ""
    plan_expires_at: str | None = None
    created_at: str
    last_login_at: str | None = None

    model_config = {"from_attributes": True}


class LoginResponse(BaseModel):
    """登录响应。"""

    token: str
    user: UserOut
