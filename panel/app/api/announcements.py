"""公告 API：列表/详情。"""
from __future__ import annotations

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.errors import BizError, NotFoundError, success
from ..models import Announcement

router = APIRouter(prefix="/api/announcements", tags=["公告"])


@router.get("")
def list_announcements(
    page: int = 1,
    page_size: int = 10,
    db: Session = Depends(get_db),
):
    """公告列表（仅显示上线状态）。"""
    q = (
        db.query(Announcement)
        .filter(Announcement.status == "active")
        .order_by(Announcement.id.desc())
    )
    total = q.count()
    items = q.offset((page - 1) * page_size).limit(page_size).all()
    return success(
        {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [
                {
                    "id": a.id,
                    "title": a.title,
                    "content": a.content[:200] + ("..." if len(a.content) > 200 else ""),
                    "author": a.author,
                    "created_at": a.created_at.isoformat() if a.created_at else None,
                }
                for a in items
            ],
        }
    )


@router.get("/{announcement_id}")
def announcement_detail(announcement_id: int, db: Session = Depends(get_db)):
    """公告详情。"""
    announcement = db.get(Announcement, announcement_id)
    if announcement is None:
        raise NotFoundError(message="公告不存在")
    if announcement.status != "active":
        raise BizError(400, 5002, "公告已下线")
    return success(
        {
            "id": announcement.id,
            "title": announcement.title,
            "content": announcement.content,
            "author": announcement.author,
            "created_at": announcement.created_at.isoformat() if announcement.created_at else None,
            "updated_at": announcement.updated_at.isoformat() if announcement.updated_at else None,
        }
    )
