"""WeaveNet 织网穿透 面板主应用入口。

负责：
- 初始化数据库与种子数据
- 注册全部 API 路由
- 挂载面板前端静态文件与官网 Jinja2 模板
- 启动后台任务
"""
from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, Request
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

from .api import agent, announcements, auth, client, domains, nodes, points, signin, stats, tickets, tunnels, user
from .api.admin import dashboard, nodes as admin_nodes, users as admin_users
from .core.config import get_settings
from .core.database import Base, SessionLocal, engine
from .core.errors import register_exception_handlers
from .core.logger import get_logger
from .core.redis_client import redis_client
from .services.mail import mail_queue
from .tasks.scheduler import scheduler_loop

settings = get_settings()
logger = get_logger()

BASE_DIR = Path(__file__).resolve().parent.parent
TEMPLATES_DIR = BASE_DIR / "templates"
WEB_DIST_DIR = BASE_DIR / "web" / "dist"
STATIC_DIR = BASE_DIR / "static"


def init_db() -> None:
    """建表并写入种子数据。"""
    from .seed import seed_defaults

    Base.metadata.create_all(bind=engine)
    with SessionLocal() as db:
        seed_defaults(db)


@asynccontextmanager
async def lifespan(app: FastAPI):  # noqa: ANN001
    """应用生命周期：启动后台任务，关闭时清理。"""
    init_db()
    task = asyncio.create_task(scheduler_loop())
    mail_queue.start()
    logger.info("WeaveNet 面板启动完成，监听 %s:%s", settings.panel_host, settings.panel_port)
    yield
    task.cancel()
    try:
        await task
    except asyncio.CancelledError:
        pass
    await mail_queue.stop()


app = FastAPI(
    title="WeaveNet 织网穿透 面板 API",
    description="用户侧、管理后台、内核联动、客户端 API",
    version="0.1.0",
    lifespan=lifespan,
    docs_url="/api/docs",
    openapi_url="/api/openapi.json",
)

register_exception_handlers(app)

# ---------- API 路由 ----------

app.include_router(auth.router)
app.include_router(user.router)
app.include_router(tunnels.router)
app.include_router(nodes.router)
app.include_router(domains.router)
app.include_router(stats.router)
app.include_router(signin.router)
app.include_router(points.router)
app.include_router(announcements.router)
app.include_router(tickets.router)
app.include_router(agent.router)
app.include_router(client.router)

# 管理后台
app.include_router(admin_users.router)
app.include_router(admin_nodes.router)
app.include_router(dashboard.router)


@app.get("/api/health")
def health():
    """存活检查。"""
    return {"code": 0, "message": "ok", "data": {"redis": "ok" if not redis_client.degraded else "degraded"}}


# ---------- 官网路由 ----------

from fastapi.templating import Jinja2Templates  # noqa: E402

templates = Jinja2Templates(directory=str(TEMPLATES_DIR))

# 官网上下文注入
from .core.database import get_db  # noqa: E402
from .models import Announcement, Node  # noqa: E402


def _site_context(request: Request) -> dict:
    """官网公共上下文。"""
    db = next(get_db())
    try:
        announcements = (
            db.query(Announcement)
            .filter(Announcement.status == "active")
            .order_by(Announcement.id.desc())
            .limit(5)
            .all()
        )
        nodes = db.query(Node).order_by(Node.id).all()
    finally:
        db.close()
    return {
        "request": request,
        "site_name": settings.app_name,
        "announcements": announcements,
        "nodes": nodes,
    }


def _render(template: str, request: Request, extra: dict | None = None) -> FileResponse:
    ctx = _site_context(request)
    if extra:
        ctx.update(extra)
    return templates.TemplateResponse(template, ctx)


@app.get("/")
def index(request: Request):
    return _render("index.html", request)


@app.get("/download")
def download(request: Request):
    return _render("download.html", request)


@app.get("/docs")
def docs(request: Request):
    return _render("docs.html", request)


@app.get("/announcements")
def announcements_page(request: Request):
    db = next(get_db())
    try:
        items = (
            db.query(Announcement)
            .filter(Announcement.status == "active")
            .order_by(Announcement.id.desc())
            .all()
        )
    finally:
        db.close()
    return _render("announcements.html", request, {"announcement_items": items})


@app.get("/announcements/{announcement_id}")
def announcement_page(announcement_id: int, request: Request):
    db = next(get_db())
    try:
        item = db.get(Announcement, announcement_id)
    finally:
        db.close()
    if item is None or item.status != "active":
        from fastapi.responses import RedirectResponse

        return RedirectResponse("/announcements")
    return _render("announcement.html", request, {"announcement": item})


@app.get("/pricing")
def pricing(request: Request):
    return _render("pricing.html", request)


@app.get("/about")
def about(request: Request):
    return _render("about.html", request)


@app.get("/terms")
def terms(request: Request):
    return _render("terms.html", request)


@app.get("/privacy")
def privacy(request: Request):
    return _render("privacy.html", request)


@app.get("/changelog")
def changelog(request: Request):
    return _render("changelog.html", request)


@app.get("/status")
def status_page(request: Request):
    return _render("status.html", request)


@app.get("/help")
def help_page(request: Request):
    return _render("help.html", request)


# 面板前端挂载（构建产物存在时）
if WEB_DIST_DIR.exists():

    from starlette.exceptions import HTTPException as StarletteHTTPException

    class SPAStaticFiles(StaticFiles):
        """SPA 静态托管：未命中文件时回退到 index.html，支持 hash 路由深链（如 /panel/register）。"""

        async def get_response(self, path: str, scope):
            try:
                return await super().get_response(path, scope)
            except StarletteHTTPException as exc:
                if exc.status_code == 404:
                    return await super().get_response("index.html", scope)
                raise

    app.mount("/panel", SPAStaticFiles(directory=str(WEB_DIST_DIR), html=True), name="panel")

# 静态资源
if STATIC_DIR.exists():
    app.mount("/static", StaticFiles(directory=str(STATIC_DIR)), name="static")
