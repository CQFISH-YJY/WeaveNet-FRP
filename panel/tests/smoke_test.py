"""面板后端快速冒烟测试：注册、验证、登录、节点、隧道。

用法：在 panel 目录执行 .venv/Scripts/python tests/smoke_test.py
"""
from __future__ import annotations

import sys
from pathlib import Path

import httpx

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

BASE = "http://127.0.0.1:8000"

failures: list[str] = []


def check(name: str, resp: httpx.Response, expect_code: int = 0) -> dict:
    if resp.status_code == 204:
        if expect_code == 0:
            print(f"[OK] {name}")
            return {}
        failures.append(f"{name}: 204 unexpected")
        return {}
    data = resp.json()
    if data.get("code") != expect_code:
        failures.append(f"{name}: code={data.get('code')} msg={data.get('message')} body={data}")
        print(f"[FAIL] {name}: {data}")
    else:
        print(f"[OK] {name}")
    return data.get("data") or {}


def main() -> None:
    import random

    suffix = random.randint(1000, 9999)
    username = f"smoke{suffix}"
    email = f"smoke{suffix}@test.com"
    password = "smokepass123"

    with httpx.Client(base_url=BASE, timeout=10) as c:
        # 注册
        r = c.post("/api/auth/register", json={"username": username, "email": email, "password": password})
        check("register", r, 0)

        # 从数据库取验证码并验证（开发环境验证码写入日志，这里直接查库）
        from app.core.database import SessionLocal
        from app.models import EmailCode, Node, User

        db = SessionLocal()
        user = db.query(User).filter(User.username == username).first()
        code_row = (
            db.query(EmailCode)
            .filter(EmailCode.email == email, EmailCode.purpose == "register")
            .order_by(EmailCode.id.desc())
            .first()
        )
        code = code_row.code if code_row else "000000"
        # 演示节点默认离线，冒烟测试置为 online 以验证隧道流程
        for node in db.query(Node).all():
            node.status = "online"
        db.commit()
        db.close()

        r = c.post("/api/auth/email-verify", json={"email": email, "code": code, "purpose": "register"})
        check("email-verify", r, 0)

        # 登录
        r = c.post("/api/auth/login", json={"username": username, "password": password})
        data = check("login", r, 0)
        token = data.get("token", "")
        headers = {"Authorization": f"Bearer {token}"}

        # 用户资料
        r = c.get("/api/user/profile", headers=headers)
        check("user profile", r, 0)

        # 节点列表
        r = c.get("/api/nodes", headers=headers)
        nodes = check("nodes list", r, 0)
        node_id = nodes[0]["id"] if nodes else None
        if node_id is None:
            print("[WARN] 无可用节点，跳过隧道测试")
        else:
            # 创建隧道
            r = c.post(
                "/api/tunnels",
                json={
                    "name": "smoke-tunnel",
                    "node_id": node_id,
                    "type": "tcp",
                    "local_ip": "127.0.0.1",
                    "local_port": 8080,
                },
                headers=headers,
            )
            tunnel = check("create tunnel", r, 0)

            # 隧道列表
            r = c.get("/api/tunnels", headers=headers)
            check("tunnel list", r, 0)

            # 生成配置
            tid = tunnel.get("id")
            r = c.post(f"/api/tunnels/{tid}/config", headers=headers)
            check("tunnel config", r, 0)

            # 签到
            r = c.post("/api/signin", headers=headers)
            check("signin", r, 0)

            # 积分流水
            r = c.get("/api/points/logs", headers=headers)
            check("points logs", r, 0)

            # 统计
            r = c.get("/api/stats/overview", headers=headers)
            check("stats overview", r, 0)

        # 登出
        r = c.post("/api/auth/logout", headers=headers)
        check("logout", r, 0)

    print()
    if failures:
        print(f"SMOKE FAILED: {len(failures)} 项失败")
        sys.exit(1)
    print("SMOKE PASSED: 全部通过")


if __name__ == "__main__":
    main()
