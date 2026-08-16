"""工单 API：创建/回复/关闭。"""
from __future__ import annotations

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_current_user
from ..core.errors import BizError, NotFoundError, success
from ..models import Ticket, User
from ..schemas.schemas import TicketCreate, TicketReply

router = APIRouter(prefix="/api/tickets", tags=["工单"])


@router.post("")
def create_ticket(
    payload: TicketCreate,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """创建工单。"""
    ticket = Ticket(
        user_id=user.id,
        title=payload.title,
        content=payload.content,
        status="open",
    )
    db.add(ticket)
    db.commit()
    db.refresh(ticket)
    return success(_serialize(ticket), "工单已提交", 201)


@router.get("")
def list_tickets(
    page: int = 1,
    page_size: int = 20,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """工单列表。"""
    q = db.query(Ticket).filter(Ticket.user_id == user.id).order_by(Ticket.id.desc())
    total = q.count()
    items = q.offset((page - 1) * page_size).limit(page_size).all()
    return success(
        {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [_serialize(t) for t in items],
        }
    )


@router.get("/{ticket_id}")
def ticket_detail(
    ticket_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """工单详情。"""
    ticket = _get_owned_ticket(db, user, ticket_id)
    return success(_serialize(ticket))


@router.post("/{ticket_id}/reply")
def reply_ticket(
    ticket_id: int,
    payload: TicketReply,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """回复工单（用户补充信息）。"""
    ticket = _get_owned_ticket(db, user, ticket_id)
    if ticket.status != "open":
        raise BizError(400, 5001, "工单已关闭，无法回复")
    ticket.content += f"\n\n【用户回复】\n{payload.content}"
    db.commit()
    return success(_serialize(ticket), "回复成功")


@router.post("/{ticket_id}/close")
def close_ticket(
    ticket_id: int,
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """关闭工单。"""
    ticket = _get_owned_ticket(db, user, ticket_id)
    if ticket.status != "open":
        raise BizError(400, 5001, "工单已关闭")
    ticket.status = "closed"
    db.commit()
    return success(_serialize(ticket), "工单已关闭")


def _serialize(ticket: Ticket) -> dict:
    return {
        "id": ticket.id,
        "user_id": ticket.user_id,
        "title": ticket.title,
        "content": ticket.content,
        "status": ticket.status,
        "admin_reply": ticket.admin_reply,
        "created_at": ticket.created_at.isoformat() if ticket.created_at else None,
        "updated_at": ticket.updated_at.isoformat() if ticket.updated_at else None,
    }


def _get_owned_ticket(db: Session, user: User, ticket_id: int) -> Ticket:
    ticket = (
        db.query(Ticket)
        .filter(Ticket.id == ticket_id, Ticket.user_id == user.id)
        .first()
    )
    if ticket is None:
        raise NotFoundError(message="工单不存在")
    return ticket
