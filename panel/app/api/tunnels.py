"""隧道 API：CRUD、启停、配置生成、状态。"""
from __future__ import annotations

import json

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import BizError, ConflictError, NotFoundError, success
from ..core.redis_client import redis_client
from ..models import Domain, Node, Tunnel, User
from ..schemas.schemas import TunnelCreate, TunnelUpdate
from ..services.frp_config import (
    allocate_remote_port,
    build_public_address,
    check_remote_port_available,
    check_subdomain_available,
    check_tunnel_limit,
    generate_frpc_config,
)

router = APIRouter(prefix="/api/tunnels", tags=["隧道"])


def _serialize_tunnel(tunnel: Tunnel) -> dict:
    """序列化隧道并附加实时状态。"""
    runtime = redis_client.get_tunnel_runtime(tunnel.id) or {}
    node = tunnel.node
    return {
        "id": tunnel.id,
        "user_id": tunnel.user_id,
        "node_id": tunnel.node_id,
        "name": tunnel.name,
        "type": tunnel.type,
        "local_ip": tunnel.local_ip,
        "local_port": tunnel.local_port,
        "remote_port": tunnel.remote_port,
        "subdomain": tunnel.subdomain,
        "custom_domain": tunnel.custom_domain,
        "kcp": tunnel.kcp,
        "encryption": tunnel.encryption,
        "compression": tunnel.compression,
        "secret_key": tunnel.secret_key,
        "load_balancers": tunnel.load_balancers,
        "status": tunnel.status,
        "status_detail": tunnel.status_detail,
        "created_at": tunnel.created_at.isoformat() if tunnel.created_at else None,
        "node": {
            "id": node.id,
            "name": node.name,
            "address": node.address,
            "port": node.port,
            "status": node.status,
        }
        if node
        else None,
        "online": bool(runtime.get("online", False)) or tunnel.status == "running",
        "connections": int(runtime.get("connections", 0)),
        "today_in": int(runtime.get("in", 0)),
        "today_out": int(runtime.get("out", 0)),
        "public_address": build_public_address(tunnel, node),
    }


@router.get("")
def list_tunnels(
    type: str | None = None,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """隧道列表。"""
    q = db.query(Tunnel).filter(Tunnel.user_id == user.id)
    if type:
        q = q.filter(Tunnel.type == type)
    tunnels = q.order_by(Tunnel.id.desc()).all()
    return success([_serialize_tunnel(t) for t in tunnels])


@router.post("")
def create_tunnel(
    payload: TunnelCreate,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """创建隧道。"""
    db.refresh(user)
    plan = user.plan
    if plan is None:
        raise BizError(400, 2001, "套餐配置异常，请联系管理员")
    check_tunnel_limit(user, plan)

    node = db.get(Node, payload.node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    if node.status != "online":
        raise BizError(400, 3004, "节点离线或维护中，请选择其他节点")

    # 远程端口：自动分配或校验指定端口
    remote_port = payload.remote_port
    if payload.type in ("tcp", "udp", "stcp", "xtcp", "kcp", "loadbalance"):
        if remote_port is None:
            remote_port = allocate_remote_port(db, node.id)
        elif not check_remote_port_available(db, node.id, remote_port):
            raise ConflictError(3001, "远程端口已被占用，请更换端口")
    else:
        # http/https 使用域名路由，无需远程端口（可自动分配占位）
        remote_port = None

    # 域名相关校验
    if payload.subdomain:
        if not check_subdomain_available(db, payload.subdomain):
            raise ConflictError(3002, "子域名已被占用，请更换")
    if payload.custom_domain:
        existed = (
            db.query(Tunnel)
            .filter(Tunnel.custom_domain == payload.custom_domain, Tunnel.user_id != user.id)
            .first()
        )
        if existed:
            raise ConflictError(3002, "自定义域名已被其他用户使用")

    tunnel = Tunnel(
        user_id=user.id,
        node_id=node.id,
        name=payload.name,
        type=payload.type,
        local_ip=payload.local_ip,
        local_port=payload.local_port,
        remote_port=remote_port,
        subdomain=payload.subdomain,
        custom_domain=payload.custom_domain,
        kcp=payload.kcp,
        encryption=payload.encryption,
        compression=payload.compression,
        secret_key=payload.secret_key,
        load_balancers=json.dumps(payload.load_balancers, ensure_ascii=False)
        if payload.load_balancers
        else None,
        status="stopped",
    )
    db.add(tunnel)
    db.commit()
    db.refresh(tunnel)
    return success(_serialize_tunnel(tunnel), "隧道创建成功", 201)


@router.get("/{tunnel_id}")
def tunnel_detail(
    tunnel_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """隧道详情。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)
    return success(_serialize_tunnel(tunnel))


@router.put("/{tunnel_id}")
def update_tunnel(
    tunnel_id: int,
    payload: TunnelUpdate,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """修改隧道。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)

    if payload.remote_port is not None and payload.remote_port != tunnel.remote_port:
        if not check_remote_port_available(db, tunnel.node_id, payload.remote_port, tunnel.id):
            raise ConflictError(3001, "远程端口已被占用，请更换端口")
        tunnel.remote_port = payload.remote_port

    if payload.subdomain is not None and payload.subdomain != tunnel.subdomain:
        if not check_subdomain_available(db, payload.subdomain):
            raise ConflictError(3002, "子域名已被占用，请更换")
        tunnel.subdomain = payload.subdomain or None

    for field in ("name", "local_ip", "local_port", "custom_domain", "kcp", "encryption", "compression", "secret_key"):
        value = getattr(payload, field, None)
        if value is not None:
            setattr(tunnel, field, value)

    if payload.load_balancers is not None:
        tunnel.load_balancers = json.dumps(payload.load_balancers, ensure_ascii=False)

    db.commit()
    db.refresh(tunnel)
    return success(_serialize_tunnel(tunnel), "隧道已更新")


@router.delete("/{tunnel_id}", status_code=204)
def delete_tunnel(
    tunnel_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """删除隧道。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)
    db.query(Domain).filter(Domain.tunnel_id == tunnel.id).update({"tunnel_id": None})
    redis_client.delete(f"tunnel:runtime:{tunnel.id}")
    db.delete(tunnel)
    db.commit()


@router.post("/{tunnel_id}/start")
def start_tunnel(
    tunnel_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """启动隧道。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)
    node = tunnel.node
    if node is None or node.status == "offline":
        raise BizError(400, 3004, "节点离线，无法启动隧道")
    if node.status == "maintenance":
        raise BizError(400, 3004, "节点维护中，请稍后再试")
    tunnel.status = "running"
    tunnel.status_detail = "已请求启动"
    db.commit()
    # 通知 frps 热更新（由 agent 每 10s 拉取，此处刷新 Redis 缓存加速感知）
    redis_client.set_json(f"tunnel:want:{tunnel.id}", 3600, {"action": "start"})
    return success(_serialize_tunnel(tunnel), "启动指令已下发")


@router.post("/{tunnel_id}/stop")
def stop_tunnel(
    tunnel_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """停止隧道。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)
    tunnel.status = "stopped"
    tunnel.status_detail = "已请求停止"
    db.commit()
    redis_client.set_json(f"tunnel:want:{tunnel.id}", 3600, {"action": "stop"})
    return success(_serialize_tunnel(tunnel), "停止指令已下发")


@router.post("/{tunnel_id}/config")
def tunnel_config(
    tunnel_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """生成 frpc 配置。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)
    node = tunnel.node
    if node is None:
        raise NotFoundError(message="节点不存在")
    config = generate_frpc_config(tunnel, user, node)
    return success({"config": config, "filename": f"frpc-{tunnel.name}.toml"})


@router.get("/{tunnel_id}/status")
def tunnel_status(
    tunnel_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """查询隧道状态。"""
    tunnel = _get_owned_tunnel(db, user, tunnel_id)
    runtime = redis_client.get_tunnel_runtime(tunnel.id) or {}
    return success(
        {
            "tunnel_id": tunnel.id,
            "status": tunnel.status,
            "status_detail": tunnel.status_detail,
            "online": bool(runtime.get("online", False)),
            "connections": int(runtime.get("connections", 0)),
            "today_in": int(runtime.get("in", 0)),
            "today_out": int(runtime.get("out", 0)),
            "updated_at": runtime.get("ts"),
        }
    )


def _get_owned_tunnel(db: Session, user: User, tunnel_id: int) -> Tunnel:
    """获取当前用户所属隧道，ORM 层强制 user_id 过滤防越权。"""
    tunnel = (
        db.query(Tunnel)
        .filter(Tunnel.id == tunnel_id, Tunnel.user_id == user.id)
        .first()
    )
    if tunnel is None:
        raise NotFoundError(message="隧道不存在")
    return tunnel
