"""pytest 配置：使用隔离的测试数据库。"""
from __future__ import annotations

import os
import sys
from pathlib import Path

# 将 panel 目录加入 sys.path
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

# 设置测试环境变量（必须在导入 app 之前）
os.environ["DATABASE_URL"] = "sqlite:///./test_weavenet.db"
os.environ["REDIS_URL"] = "redis://127.0.0.1:6379/15"
os.environ["WEAVE_ENV"] = "test"
os.environ["RATE_LIMIT_ENABLED"] = "false"

import pytest  # noqa: E402

from fastapi.testclient import TestClient  # noqa: E402

# 每次会话使用全新测试数据库，避免跨运行状态污染（如管理员改过系统配置后残留）
for _suffix in ("", "-shm", "-wal"):
    _p = Path(f"test_weavenet.db{_suffix}")
    if _p.exists():
        _p.unlink()


@pytest.fixture(scope="session")
def client():
    """提供测试客户端，会话级。"""
    from app.main import app

    with TestClient(app) as c:
        yield c


@pytest.fixture()
def test_user(client):
    """创建一个已验证用户并返回 token。"""
    import random

    from app.core.database import SessionLocal
    from app.models import EmailCode, Node, User

    suffix = random.randint(10000, 99999)
    username = f"apitest{suffix}"
    email = f"apitest{suffix}@test.com"

    # 注册
    r = client.post("/api/auth/register", json={
        "username": username,
        "email": email,
        "password": "testpass123",
    })
    assert r.status_code in (200, 201)

    # 取验证码并验证
    db = SessionLocal()
    code_row = (
        db.query(EmailCode)
        .filter(EmailCode.email == email, EmailCode.purpose == "register")
        .order_by(EmailCode.id.desc())
        .first()
    )
    code = code_row.code if code_row else "000000"
    # 演示节点置为 online 以便创建隧道
    for node in db.query(Node).all():
        node.status = "online"
    db.commit()
    db.close()

    r = client.post("/api/auth/email-verify", json={"email": email, "code": code, "purpose": "register"})
    assert r.status_code == 200

    # 登录
    r = client.post("/api/auth/login", json={"username": username, "password": "testpass123"})
    assert r.status_code == 200
    data = r.json()["data"]
    return {
        "username": username,
        "email": email,
        "password": "testpass123",
        "token": data["token"],
        "headers": {"Authorization": f"Bearer {data['token']}"},
    }


@pytest.fixture()
def admin_headers(client):
    """管理员登录。"""
    r = client.post("/api/auth/login", json={"username": "admin", "password": "admin123"})
    assert r.status_code == 200
    token = r.json()["data"]["token"]
    return {"Authorization": f"Bearer {token}"}
