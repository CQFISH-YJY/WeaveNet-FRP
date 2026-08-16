"""隧道、节点、域名、套餐、统计、工单等业务请求/响应模型。"""
from __future__ import annotations

from pydantic import BaseModel, Field

TUNNEL_TYPES = (
    "tcp",
    "udp",
    "http",
    "https",
    "stcp",
    "xtcp",
    "kcp",
    "loadbalance",
)


# ---------- 节点 ----------

class NodeOut(BaseModel):
    """节点信息响应。"""

    id: int
    name: str
    address: str
    port: int
    status: str
    speed_limit_mbps: int
    remark: str
    last_heartbeat_at: str | None = None

    model_config = {"from_attributes": True}


# ---------- 隧道 ----------

class TunnelCreate(BaseModel):
    """创建隧道请求。"""

    name: str = Field(min_length=1, max_length=64)
    node_id: int
    type: str = Field(pattern=r"^(tcp|udp|http|https|stcp|xtcp|kcp|loadbalance)$")
    local_ip: str = Field(default="127.0.0.1", max_length=255)
    local_port: int = Field(ge=1, le=65535)
    remote_port: int | None = Field(default=None, ge=1, le=65535)
    subdomain: str | None = Field(default=None, max_length=128)
    custom_domain: str | None = Field(default=None, max_length=255)
    kcp: bool = False
    encryption: bool = True
    compression: bool = False
    secret_key: str | None = Field(default=None, max_length=128)
    load_balancers: list[dict] | None = None


class TunnelUpdate(BaseModel):
    """更新隧道请求。"""

    name: str | None = Field(default=None, min_length=1, max_length=64)
    local_ip: str | None = Field(default=None, max_length=255)
    local_port: int | None = Field(default=None, ge=1, le=65535)
    remote_port: int | None = Field(default=None, ge=1, le=65535)
    subdomain: str | None = Field(default=None, max_length=128)
    custom_domain: str | None = Field(default=None, max_length=255)
    kcp: bool | None = None
    encryption: bool | None = None
    compression: bool | None = None
    secret_key: str | None = Field(default=None, max_length=128)
    load_balancers: list[dict] | None = None


class TunnelOut(BaseModel):
    """隧道信息响应。"""

    id: int
    user_id: int
    node_id: int
    name: str
    type: str
    local_ip: str
    local_port: int
    remote_port: int | None
    subdomain: str | None
    custom_domain: str | None
    kcp: bool
    encryption: bool
    compression: bool
    secret_key: str | None
    load_balancers: str | None
    status: str
    status_detail: str
    created_at: str
    node: NodeOut | None = None
    # 实时补充字段
    online: bool = False
    connections: int = 0
    today_in: int = 0
    today_out: int = 0
    public_address: str = ""

    model_config = {"from_attributes": True}


# ---------- 域名 ----------

class DomainCreate(BaseModel):
    """申请免费域名请求。"""

    subdomain: str = Field(min_length=3, max_length=128, pattern=r"^[a-zA-Z0-9\-]+$")
    tunnel_id: int | None = None


class DomainOut(BaseModel):
    """域名信息响应。"""

    id: int
    user_id: int
    tunnel_id: int | None
    subdomain: str
    full_domain: str
    status: str
    created_at: str

    model_config = {"from_attributes": True}


# ---------- 套餐 ----------

class PlanOut(BaseModel):
    """套餐信息响应。"""

    id: int
    name: str
    speed_limit_mbps: int
    tunnel_limit: int
    domain_limit: int
    sort: int

    model_config = {"from_attributes": True}


class PlanUpdate(BaseModel):
    """修改套餐请求。"""

    name: str | None = Field(default=None, max_length=64)
    speed_limit_mbps: int | None = Field(default=None, ge=1)
    tunnel_limit: int | None = Field(default=None, ge=0)
    domain_limit: int | None = Field(default=None, ge=0)
    sort: int | None = None


# ---------- 工单 ----------

class TicketCreate(BaseModel):
    """创建工单请求。"""

    title: str = Field(min_length=2, max_length=128)
    content: str = Field(min_length=5, max_length=5000)


class TicketReply(BaseModel):
    """回复工单请求。"""

    content: str = Field(min_length=1, max_length=5000)


class TicketOut(BaseModel):
    """工单信息响应。"""

    id: int
    user_id: int
    title: str
    content: str
    status: str
    admin_reply: str | None
    created_at: str
    updated_at: str

    model_config = {"from_attributes": True}


# ---------- 公告 ----------

class AnnouncementCreate(BaseModel):
    """发布公告请求。"""

    title: str = Field(min_length=2, max_length=128)
    content: str = Field(min_length=5, max_length=20000)
    author: str = Field(min_length=1, max_length=64)


class AnnouncementUpdate(BaseModel):
    """修改公告请求。"""

    title: str | None = Field(default=None, min_length=2, max_length=128)
    content: str | None = Field(default=None, min_length=5, max_length=20000)
    author: str | None = Field(default=None, min_length=1, max_length=64)


class AnnouncementOut(BaseModel):
    """公告信息响应。"""

    id: int
    title: str
    content: str
    author: str
    status: str
    created_at: str
    updated_at: str

    model_config = {"from_attributes": True}


# ---------- 用户管理（管理端） ----------

class AdminUserUpdate(BaseModel):
    """管理员调整用户请求。"""

    plan_id: int | None = None
    points: int | None = None


class AdminUserBan(BaseModel):
    """封禁用户请求。"""

    reason: str = Field(default="", max_length=255)


# ---------- 节点管理（管理端） ----------

class NodeCreate(BaseModel):
    """新增节点请求。"""

    name: str = Field(min_length=1, max_length=64)
    address: str = Field(min_length=1, max_length=255)
    port: int = Field(default=7000, ge=1, le=65535)
    speed_limit_mbps: int = Field(default=100, ge=1)
    remark: str = Field(default="", max_length=512)


class NodeUpdate(BaseModel):
    """修改节点请求。"""

    name: str | None = Field(default=None, min_length=1, max_length=64)
    address: str | None = Field(default=None, min_length=1, max_length=255)
    port: int | None = Field(default=None, ge=1, le=65535)
    speed_limit_mbps: int | None = Field(default=None, ge=1)
    remark: str | None = Field(default=None, max_length=512)


# ---------- 系统配置 ----------

class ConfigUpdate(BaseModel):
    """修改系统配置请求。"""

    key: str = Field(min_length=1, max_length=64)
    value: str


# ---------- 通用分页 ----------

class PageResult(BaseModel):
    """通用分页响应。"""

    total: int
    page: int
    page_size: int
    items: list
