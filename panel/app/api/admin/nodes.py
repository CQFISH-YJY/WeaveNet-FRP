"""管理后台：节点管理 API。"""
from __future__ import annotations

from fastapi import APIRouter, Depends, Request
from sqlalchemy.orm import Session

from ...core.database import get_db
from ...core.deps import get_current_admin
from ...core.errors import ConflictError, NotFoundError, success
from ...core.security import generate_agent_token
from ...models import Node, OperationLog, User
from ...schemas.schemas import NodeCreate, NodeUpdate

router = APIRouter(prefix="/api/admin/nodes", tags=["管理-节点"])


def _log(db: Session, request: Request, admin: User, action: str, target_type: str, target_id: int | None, detail: str) -> None:  # noqa: ANN001
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


def _serialize(node: Node) -> dict:
    return {
        "id": node.id,
        "name": node.name,
        "address": node.address,
        "port": node.port,
        "status": node.status,
        "speed_limit_mbps": node.speed_limit_mbps,
        "agent_token": node.agent_token,
        "remark": node.remark,
        "last_heartbeat_at": node.last_heartbeat_at.isoformat() if node.last_heartbeat_at else None,
        "created_at": node.created_at.isoformat() if node.created_at else None,
    }


@router.get("")
def list_nodes(admin: User = Depends(get_current_admin), db: Session = Depends(get_db)):
    """节点列表。"""
    nodes = db.query(Node).order_by(Node.id).all()
    return success([_serialize(n) for n in nodes])


@router.post("")
def create_node(
    payload: NodeCreate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """新增节点。"""
    if db.query(Node).filter(Node.name == payload.name).first():
        raise ConflictError(0, "节点名称已存在")
    node = Node(
        name=payload.name,
        address=payload.address,
        port=payload.port,
        status="offline",
        speed_limit_mbps=payload.speed_limit_mbps,
        agent_token=generate_agent_token(),
        remark=payload.remark,
    )
    db.add(node)
    _log(db, request, admin, "create_node", "node", None, f"新增节点 {payload.name}")
    db.commit()
    db.refresh(node)
    return success(_serialize(node), "节点创建成功", 201)


@router.put("/{node_id}")
def update_node(
    node_id: int,
    payload: NodeUpdate,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """修改节点。"""
    node = db.get(Node, node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    for field in ("name", "address", "port", "speed_limit_mbps", "remark"):
        value = getattr(payload, field, None)
        if value is not None:
            setattr(node, field, value)
    _log(db, request, admin, "update_node", "node", node.id, f"修改节点 {node.name}")
    db.commit()
    db.refresh(node)
    return success(_serialize(node), "节点已更新")


@router.delete("/{node_id}", status_code=204)
def delete_node(
    node_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """删除节点。"""
    node = db.get(Node, node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    _log(db, request, admin, "delete_node", "node", node.id, f"删除节点 {node.name}")
    db.delete(node)
    db.commit()


@router.post("/{node_id}/start")
def start_node(
    node_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """启用节点。"""
    node = db.get(Node, node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    node.status = "online" if node.status == "maintenance" else node.status
    if node.status == "offline":
        # 离线状态启用后等待心跳确认，先置为 online 占位
        node.status = "online"
    _log(db, request, admin, "start_node", "node", node.id, f"启用节点 {node.name}")
    db.commit()
    return success(_serialize(node), "节点已启用")


@router.post("/{node_id}/stop")
def stop_node(
    node_id: int,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """停用节点。"""
    node = db.get(Node, node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    node.status = "maintenance"
    _log(db, request, admin, "stop_node", "node", node.id, f"停用节点 {node.name}")
    db.commit()
    return success(_serialize(node), "节点已停用")


@router.put("/{node_id}/speed")
def node_speed(
    node_id: int,
    payload: dict,
    request: Request,
    admin: User = Depends(get_current_admin),
    db: Session = Depends(get_db),
):
    """节点限速配置。"""
    node = db.get(Node, node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    speed = payload.get("speed_limit_mbps")
    if speed is None or not isinstance(speed, int) or speed < 1:
        raise NotFoundError(0, "限速值必须为正整数")
    node.speed_limit_mbps = speed
    _log(db, request, admin, "update_node_speed", "node", node.id, f"节点限速调整为 {speed}Mbps")
    db.commit()
    return success(_serialize(node), "节点限速已更新")
