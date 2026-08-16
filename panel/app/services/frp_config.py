"""frpc 配置生成与隧道辅助服务。"""
from __future__ import annotations

import ipaddress
import json
from typing import Any

from ..core.config import get_settings
from ..core.errors import BizError, ConflictError, NotFoundError
from ..models import Domain, Node, Plan, Tunnel, User

settings = get_settings()

# 系统保留端口与常用占用端口，禁止分配给隧道
RESERVED_PORTS = set(range(1, 1024))
RESERVED_PORTS.update({7000, 8000, 9001, 5432, 3306, 6379, 27017})


def check_tunnel_limit(user: User, plan: Plan) -> None:
    """校验隧道数量额度。"""
    from ..models import Tunnel

    count = len(user.tunnels)
    if count >= plan.tunnel_limit:
        raise BizError(400, 2001, f"套餐最多可创建 {plan.tunnel_limit} 条隧道，请升级套餐或删除不再使用的隧道")


def check_domain_limit(user: User, plan: Plan) -> None:
    """校验免费域名数量额度。"""
    from ..models import Domain

    count = (
        user.id
        and __count_active_domains(user.id)
    )
    if count >= plan.domain_limit:
        raise BizError(400, 3003, f"套餐最多可申请 {plan.domain_limit} 个免费域名")


def __count_active_domains(user_id: int) -> int:
    from ..core.database import SessionLocal

    with SessionLocal() as db:
        from ..models import Domain

        return (
            db.query(Domain)
            .filter(Domain.user_id == user_id, Domain.status == "active")
            .count()
        )


def allocate_remote_port(db, node_id: int) -> int:
    """从节点端口池分配一个空闲远程端口。

    端口池：1024-65535 中的常用端口段（20000-39999 优先），
    排除保留端口与已被其他隧道占用的端口。
    """
    from ..models import Tunnel

    occupied = {
        port
        for (port,) in db.query(Tunnel.remote_port)
        .filter(Tunnel.node_id == node_id, Tunnel.remote_port.isnot(None))
        .all()
        if port
    }
    # 优先使用 20000-39999，避免撞上常见服务
    for start in (20000, 40000, 1024, 10000):
        for port in range(start, start + 1000):
            if port in RESERVED_PORTS or port in occupied:
                continue
            return port
    raise BizError(409, 3001, "节点远程端口已用尽，请联系管理员")


def check_remote_port_available(db, node_id: int, port: int, exclude_tunnel_id: int | None = None) -> bool:
    """校验远程端口是否可用。"""
    from ..models import Tunnel

    if port in RESERVED_PORTS:
        return False
    q = db.query(Tunnel).filter(Tunnel.node_id == node_id, Tunnel.remote_port == port)
    if exclude_tunnel_id:
        q = q.filter(Tunnel.id != exclude_tunnel_id)
    return q.first() is None


def check_subdomain_available(db, subdomain: str, exclude_domain_id: int | None = None) -> bool:
    """校验子域名是否被占用。"""
    q = db.query(Domain).filter(Domain.subdomain == subdomain, Domain.status == "active")
    if exclude_domain_id:
        q = q.filter(Domain.id != exclude_domain_id)
    return q.first() is None


def build_public_address(tunnel: Tunnel, node: Node | None = None) -> str:
    """构造公网访问地址。"""
    node = node or tunnel.node
    host = node.address if node else settings.panel_base_url
    if tunnel.type in ("http", "https"):
        domain = tunnel.custom_domain or tunnel.subdomain
        if domain:
            scheme = "https" if tunnel.type == "https" else "http"
            return f"{scheme}://{domain}"
        return f"{tunnel.type}://{host}:{tunnel.remote_port or ''}"
    if tunnel.remote_port:
        return f"{host}:{tunnel.remote_port}"
    return host


def generate_frpc_config(tunnel: Tunnel, user: User, node: Node) -> str:
    """生成单条隧道的 frpc.toml 配置。"""
    lines: list[str] = []
    lines.append("serverAddr = %r" % node.address)
    lines.append("serverPort = %d" % node.port)
    lines.append("")
    lines.append("auth.method = \"token\"")
    lines.append("auth.token = %r" % user_token_for_frpc(user))

    name = f"{tunnel.name}-{tunnel.id}"
    lines.append("")
    lines.append(f"[[proxies]]")
    lines.append(f"name = {name!r}")
    lines.append(f"type = {tunnel.type!r}")
    lines.append(f"localIP = {tunnel.local_ip!r}")
    lines.append(f"localPort = {tunnel.local_port}")
    if tunnel.remote_port:
        lines.append(f"remotePort = {tunnel.remote_port}")
    if tunnel.subdomain:
        lines.append(f"subdomain = {tunnel.subdomain!r}")
    if tunnel.custom_domain:
        lines.append(f"customDomains = [{tunnel.custom_domain!r}]")
    lines.append(f"transport.kcp = {'true' if tunnel.kcp else 'false'}")
    lines.append(f"transport.bandwidthLimit = {tunnel_bandwidth(user)}")
    lines.append(f"transport.bandwidthLimitMode = \"client\"")
    if tunnel.encryption:
        lines.append("transport.encryption.enable = true")
    else:
        lines.append("transport.encryption.enable = false")
    if tunnel.compression:
        lines.append("transport.compression.enable = true")
    if tunnel.secret_key:
        lines.append(f"secretKey = {tunnel.secret_key!r}")
    if tunnel.type == "loadbalance" and tunnel.load_balancers:
        lbs = json.loads(tunnel.load_balancers)
        lines.append("loadBalancer.group = \"weavenet-lb\"")
        lines.append(f"loadBalancer.groupKey = {str(tunnel.id)!r}")
        for i, lb in enumerate(lbs):
            lines.append("")
            lines.append(f"[[proxies]]")
            lines.append(f"name = {f'{tunnel.name}-lb{i}-{tunnel.id}'!r}")
            lines.append("type = \"tcp\"")
            lines.append(f"localIP = {lb.get('ip', '127.0.0.1')!r}")
            lines.append(f"localPort = {lb.get('port', 80)}")
            lines.append(f"remotePort = {tunnel.remote_port}")
            lines.append("loadBalancer.group = \"weavenet-lb\"")
            lines.append(f"loadBalancer.groupKey = {str(tunnel.id)!r}")

    return "\n".join(lines)


def tunnel_bandwidth(user: User) -> str:
    """按用户套餐返回限速值（Kbps）。"""
    plan = user.plan
    mbps = plan.speed_limit_mbps if plan else 8
    return f"{mbps * 1000}KB"


def user_token_for_frpc(user: User) -> str:
    """生成 frpc 连接鉴权 Token：基于用户 ID 与全局密钥的稳定签名。"""
    import hashlib

    raw = f"{user.id}:{user.password_hash}:{settings.secret_key}".encode()
    return "u_" + hashlib.sha256(raw).hexdigest()
