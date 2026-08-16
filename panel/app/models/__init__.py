"""SQLAlchemy 数据模型。

依据设计文档第 4 章数据库模型实现 15 张核心表。
"""
from __future__ import annotations

from datetime import datetime

from sqlalchemy import (
    BigInteger,
    Boolean,
    DateTime,
    Float,
    ForeignKey,
    Integer,
    String,
    Text,
    func,
)
from sqlalchemy.orm import Mapped, mapped_column, relationship

from ..core.database import Base


class User(Base):
    """用户表。"""

    __tablename__ = "users"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    username: Mapped[str] = mapped_column(String(64), unique=True, index=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    password_hash: Mapped[str] = mapped_column(String(255))
    email_verified: Mapped[bool] = mapped_column(Boolean, default=False)
    status: Mapped[str] = mapped_column(String(16), default="active")  # active / banned
    plan_id: Mapped[int] = mapped_column(
        ForeignKey("plans.id", ondelete="RESTRICT"), default=1, index=True
    )
    plan_expires_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    points: Mapped[int] = mapped_column(Integer, default=0)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )
    last_login_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)

    plan: Mapped["Plan"] = relationship("Plan", lazy="joined")
    tunnels: Mapped[list["Tunnel"]] = relationship(
        "Tunnel", back_populates="user", cascade="all, delete-orphan"
    )


class Plan(Base):
    """套餐表。"""

    __tablename__ = "plans"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(String(64), unique=True)
    speed_limit_mbps: Mapped[int] = mapped_column(Integer, default=8)
    tunnel_limit: Mapped[int] = mapped_column(Integer, default=3)
    domain_limit: Mapped[int] = mapped_column(Integer, default=1)
    sort: Mapped[int] = mapped_column(Integer, default=0)
    # 是否可被管理员编辑（免费版仅可查看）
    is_default: Mapped[bool] = mapped_column(Boolean, default=False)


class Node(Base):
    """穿透节点表。"""

    __tablename__ = "nodes"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(String(64), unique=True)
    address: Mapped[str] = mapped_column(String(255))
    port: Mapped[int] = mapped_column(Integer, default=7000)
    status: Mapped[str] = mapped_column(String(16), default="offline")  # online/offline/maintenance
    speed_limit_mbps: Mapped[int] = mapped_column(Integer, default=100)
    agent_token: Mapped[str] = mapped_column(String(128), default="")
    remark: Mapped[str] = mapped_column(String(512), default="")
    last_heartbeat_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, server_default=func.now())

    tunnels: Mapped[list["Tunnel"]] = relationship("Tunnel", back_populates="node")


class Tunnel(Base):
    """隧道表。"""

    __tablename__ = "tunnels"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    node_id: Mapped[int] = mapped_column(
        ForeignKey("nodes.id", ondelete="RESTRICT"), index=True
    )
    name: Mapped[str] = mapped_column(String(64))
    # tcp / udp / http / https / stcp / xtcp / kcp / loadbalance
    type: Mapped[str] = mapped_column(String(20))
    local_ip: Mapped[str] = mapped_column(String(255), default="127.0.0.1")
    local_port: Mapped[int] = mapped_column(Integer, default=0)
    remote_port: Mapped[int] = mapped_column(Integer, nullable=True, index=True)
    subdomain: Mapped[str | None] = mapped_column(String(128), nullable=True, index=True)
    custom_domain: Mapped[str | None] = mapped_column(String(255), nullable=True)
    kcp: Mapped[bool] = mapped_column(Boolean, default=False)
    encryption: Mapped[bool] = mapped_column(Boolean, default=True)
    compression: Mapped[bool] = mapped_column(Boolean, default=False)
    # 访问密钥（stcp/xtcp 使用）
    secret_key: Mapped[str | None] = mapped_column(String(128), nullable=True)
    # 负载均衡：后端列表 JSON，如 [{"ip":"127.0.0.1","port":80}]
    load_balancers: Mapped[str | None] = mapped_column(Text, nullable=True)
    # 在线状态（实时状态优先取 Redis，此处为持久化兜底）
    status: Mapped[str] = mapped_column(String(16), default="stopped")  # running/stopped
    status_detail: Mapped[str] = mapped_column(String(512), default="")
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )

    user: Mapped["User"] = relationship("User", back_populates="tunnels")
    node: Mapped["Node"] = relationship("Node", back_populates="tunnels")


class Domain(Base):
    """免费域名表。"""

    __tablename__ = "domains"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    tunnel_id: Mapped[int | None] = mapped_column(
        ForeignKey("tunnels.id", ondelete="SET NULL"), nullable=True
    )
    subdomain: Mapped[str] = mapped_column(String(128), unique=True, index=True)
    full_domain: Mapped[str] = mapped_column(String(255), index=True)
    status: Mapped[str] = mapped_column(String(16), default="active")  # active/released
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )


class SigninLog(Base):
    """签到记录表。"""

    __tablename__ = "signin_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    signin_date: Mapped[str] = mapped_column(String(16), index=True)  # YYYY-MM-DD
    points: Mapped[int] = mapped_column(Integer, default=0)
    continuous_days: Mapped[int] = mapped_column(Integer, default=1)
    created_at: Mapped[datetime] = mapped_column(DateTime, server_default=func.now())


class PointsLog(Base):
    """积分流水表。"""

    __tablename__ = "points_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    change: Mapped[int] = mapped_column(Integer)  # 正数增加，负数扣除
    balance: Mapped[int] = mapped_column(Integer, default=0)  # 变动后余额
    reason: Mapped[str] = mapped_column(String(128))
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )


class Ticket(Base):
    """工单表。"""

    __tablename__ = "tickets"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    title: Mapped[str] = mapped_column(String(128))
    content: Mapped[str] = mapped_column(Text)
    status: Mapped[str] = mapped_column(String(16), default="open")  # open/closed
    admin_reply: Mapped[str | None] = mapped_column(Text, nullable=True)
    admin_id: Mapped[int | None] = mapped_column(Integer, nullable=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), onupdate=func.now()
    )


class Announcement(Base):
    """公告表。"""

    __tablename__ = "announcements"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    title: Mapped[str] = mapped_column(String(128))
    content: Mapped[str] = mapped_column(Text)
    author: Mapped[str] = mapped_column(String(64), default="")  # 发布人/部门名称
    status: Mapped[str] = mapped_column(String(16), default="active")  # active/offline
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), onupdate=func.now()
    )


class TrafficStat(Base):
    """流量统计表（按日聚合）。"""

    __tablename__ = "traffic_stats"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    tunnel_id: Mapped[int] = mapped_column(
        ForeignKey("tunnels.id", ondelete="CASCADE"), index=True
    )
    date: Mapped[str] = mapped_column(String(16), index=True)  # YYYY-MM-DD
    in_bytes: Mapped[int] = mapped_column(BigInteger, default=0)
    out_bytes: Mapped[int] = mapped_column(BigInteger, default=0)


class SystemConfig(Base):
    """系统配置表。"""

    __tablename__ = "system_configs"

    key: Mapped[str] = mapped_column(String(64), primary_key=True)
    value: Mapped[str] = mapped_column(Text, default="")


class Session(Base):
    """会话表。"""

    __tablename__ = "sessions"

    token: Mapped[str] = mapped_column(String(128), primary_key=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    expires_at: Mapped[datetime] = mapped_column(DateTime, index=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, server_default=func.now())


class EmailCode(Base):
    """邮箱验证码表。"""

    __tablename__ = "email_codes"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    email: Mapped[str] = mapped_column(String(255), index=True)
    code: Mapped[str] = mapped_column(String(16))
    purpose: Mapped[str] = mapped_column(String(32))  # register / reset_password / change_email
    expires_at: Mapped[datetime] = mapped_column(DateTime, index=True)
    used: Mapped[bool] = mapped_column(Boolean, default=False)
    attempts: Mapped[int] = mapped_column(Integer, default=0)
    created_at: Mapped[datetime] = mapped_column(DateTime, server_default=func.now())


class OperationLog(Base):
    """操作日志表（管理后台审计）。"""

    __tablename__ = "operation_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    admin_id: Mapped[int | None] = mapped_column(Integer, nullable=True, index=True)
    admin_name: Mapped[str] = mapped_column(String(64), default="")
    action: Mapped[str] = mapped_column(String(64))
    target_type: Mapped[str] = mapped_column(String(32), default="")
    target_id: Mapped[int | None] = mapped_column(Integer, nullable=True)
    detail: Mapped[str] = mapped_column(Text, default="")
    ip: Mapped[str] = mapped_column(String(64), default="")
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )


class UserPlanLog(Base):
    """套餐变更记录表。"""

    __tablename__ = "user_plan_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        ForeignKey("users.id", ondelete="CASCADE"), index=True
    )
    plan_id: Mapped[int] = mapped_column(Integer, index=True)
    plan_name: Mapped[str] = mapped_column(String(64), default="")
    reason: Mapped[str] = mapped_column(String(128), default="")
    created_at: Mapped[datetime] = mapped_column(
        DateTime, server_default=func.now(), index=True
    )
