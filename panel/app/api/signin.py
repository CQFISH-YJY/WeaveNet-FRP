"""签到 API。"""
from __future__ import annotations

from fastapi import APIRouter, Depends, Request
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..core.deps import get_current_user, require_rate_limit
from ..core.errors import success
from ..models import User
from ..services.points import can_signin, do_signin

router = APIRouter(prefix="/api/signin", tags=["签到"])


@router.post("")
def signin(
    request: Request,
    _rl=Depends(require_rate_limit("signin", limit=3, window=60)),
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """每日签到。"""
    result = do_signin(db, user)
    return success(result, "签到成功")


@router.get("/status")
def signin_status(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """签到状态。"""
    return success(can_signin(db, user))
