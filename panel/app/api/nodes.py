"""节点 API：节点列表与状态（用户侧）。"""
from __future__ import annotations

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import NotFoundError, success
from ..models import Node

router = APIRouter(prefix="/api/nodes", tags=["节点"])


@router.get("")
def list_nodes(user=Depends(get_current_user), db: Session = Depends(get_db)):
    """节点列表。"""
    nodes = db.query(Node).order_by(Node.id).all()
    return success(
        [
            {
                "id": n.id,
                "name": n.name,
                "address": n.address,
                "port": n.port,
                "status": n.status,
                "speed_limit_mbps": n.speed_limit_mbps,
                "remark": n.remark,
                "last_heartbeat_at": n.last_heartbeat_at.isoformat() if n.last_heartbeat_at else None,
            }
            for n in nodes
        ]
    )


@router.get("/{node_id}/status")
def node_status(node_id: int, user=Depends(get_current_user), db: Session = Depends(get_db)):
    """节点状态。"""
    node = db.get(Node, node_id)
    if node is None:
        raise NotFoundError(message="节点不存在")
    return success(
        {
            "id": node.id,
            "name": node.name,
            "status": node.status,
            "speed_limit_mbps": node.speed_limit_mbps,
            "last_heartbeat_at": node.last_heartbeat_at.isoformat() if node.last_heartbeat_at else None,
        }
    )
