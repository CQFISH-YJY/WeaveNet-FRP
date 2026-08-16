"""内核联动 API：frps 节点注册、心跳、配置拉取。

frps 使用独立 Agent Token 鉴权，仅允许访问本分组。
"""
from __future__ import annotations

from datetime import datetime

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_agent_node
from ..core.errors import ConflictError, NotFoundError, success
from ..core.redis_client import redis_client
from ..models import Domain, Node, TrafficStat, Tunnel, User
from ..services.frp_config import user_token_for_frpc

router = APIRouter(prefix="/api/agent", tags=["内核联动"])


class AgentRegisterRequest:
    pass


@router.post("/register")
def agent_register(
    payload: dict,
    db: Session = Depends(get_db),
):
    """节点注册。

    frps 启动时携带 Agent Token 调用；若节点不存在则自动注册。
    """
    token = payload.get("agent_token", "")
    if not token:
        raise NotFoundError(0, "缺少 agent_token")

    node = db.query(Node).filter(Node.agent_token == token).first()
    if node is None:
        # 允许按名称注册：name + token 组合
        node = db.query(Node).filter(Node.name == payload.get("name", "")).first()
        if node is None:
            raise NotFoundError(0, "节点未授权，请在管理后台创建节点后配置 Agent Token")
        node.agent_token = token

    node.address = payload.get("address", node.address)
    node.port = payload.get("port", node.port)
    node.last_heartbeat_at = datetime.now()
    node.status = "online"
    redis_client.set_node_online(node.id)
    db.commit()
    return success({"node_id": node.id, "status": "registered"})


@router.post("/heartbeat")
def agent_heartbeat(
    payload: dict,
    node: Node = Depends(get_agent_node),
    db: Session = Depends(get_db),
):
    """心跳上报：隧道状态、连接数、流量增量。"""
    node.last_heartbeat_at = datetime.now()
    node.status = "online"
    redis_client.set_node_online(node.id)

    tunnels = payload.get("tunnels") or []
    for item in tunnels:
        tunnel_id = item.get("tunnel_id")
        if not tunnel_id:
            continue
        online = bool(item.get("online", False))
        connections = int(item.get("connections", 0))
        in_delta = int(item.get("in_delta", 0))
        out_delta = int(item.get("out_delta", 0))

        # 实时状态入 Redis
        redis_client.set_tunnel_runtime(
            tunnel_id,
            {
                "online": online,
                "connections": connections,
                "in": redis_client.get_tunnel_runtime(tunnel_id).get("in", 0) + in_delta
                if redis_client.get_tunnel_runtime(tunnel_id)
                else in_delta,
                "out": redis_client.get_tunnel_runtime(tunnel_id).get("out", 0) + out_delta
                if redis_client.get_tunnel_runtime(tunnel_id)
                else out_delta,
                "ts": datetime.now().isoformat(),
            },
        )
        # 流量增量累加当日总量
        if in_delta or out_delta:
            redis_client.incr_traffic(tunnel_id, in_delta, out_delta)

        # 同步持久化状态
        tunnel = db.get(Tunnel, tunnel_id)
        if tunnel:
            tunnel.status = "running" if online else tunnel.status
            tunnel.status_detail = "在线" if online else tunnel.status_detail

    db.commit()
    return success(message="ok")


@router.get("/tunnels")
def agent_tunnels(
    node: Node = Depends(get_agent_node),
    db: Session = Depends(get_db),
):
    """拉取本节点隧道与用户限速配置（frps 每 10s 轮询）。"""
    tunnels = (
        db.query(Tunnel)
        .filter(Tunnel.node_id == node.id, Tunnel.status == "running")
        .all()
    )
    result = []
    for t in tunnels:
        user = db.get(User, t.user_id)
        if user is None:
            continue
        result.append(
            {
                "tunnel_id": t.id,
                "user_token": user_token_for_frpc(user),
                "username": user.username,
                "type": t.type,
                "name": t.name,
                "local_ip": t.local_ip,
                "local_port": t.local_port,
                "remote_port": t.remote_port,
                "subdomain": t.subdomain,
                "custom_domain": t.custom_domain,
                "kcp": t.kcp,
                "encryption": t.encryption,
                "compression": t.compression,
                "secret_key": t.secret_key,
                "bandwidth_limit_kbps": user.plan.speed_limit_mbps * 1000
                if user.plan
                else 8000,
                "load_balancers": t.load_balancers,
            }
        )
    # 同时下发所有在线用户的限速配置
    users = (
        db.query(User)
        .join(Tunnel, Tunnel.user_id == User.id)
        .filter(Tunnel.node_id == node.id, Tunnel.status == "running")
        .distinct()
        .all()
    )
    speed_limits = [
        {
            "user_token": user_token_for_frpc(u),
            "username": u.username,
            "bandwidth_limit_kbps": u.plan.speed_limit_mbps * 1000 if u.plan else 8000,
            "status": u.status,
        }
        for u in users
    ]
    # 域名路由表
    domains = (
        db.query(Domain)
        .filter(Domain.status == "active")
        .all()
    )
    return success(
        {
            "tunnels": result,
            "speed_limits": speed_limits,
            "domains": [
                {"full_domain": d.full_domain, "subdomain": d.subdomain, "tunnel_id": d.tunnel_id}
                for d in domains
            ],
        }
    )
