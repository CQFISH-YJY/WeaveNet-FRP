"""接口测试：覆盖全部 API 分组，包含鉴权与越权用例。

运行：panel 目录下 .venv/Scripts/python -m pytest tests/ -v
"""
from __future__ import annotations

import random

import pytest


class TestAuth:
    """认证 API。"""

    def test_register_login_logout(self, client):
        suffix = random.randint(10000, 99999)
        username = f"auth{suffix}"
        email = f"auth{suffix}@test.com"
        # 注册
        r = client.post("/api/auth/register", json={"username": username, "email": email, "password": "testpass123"})
        assert r.status_code in (200, 201)
        # 重复注册用户名
        r = client.post("/api/auth/register", json={"username": username, "email": f"other{suffix}@test.com", "password": "testpass123"})
        assert r.json()["business_code"] == 1004
        # 未验证邮箱登录被拒
        r = client.post("/api/auth/login", json={"username": username, "password": "testpass123"})
        assert r.json()["business_code"] == 1005
        # 错误密码
        r = client.post("/api/auth/login", json={"username": "admin", "password": "wrong"})
        assert r.status_code == 401

    def test_unauthorized_access(self, client):
        """未登录访问受保护接口返回 401。"""
        r = client.get("/api/user/profile")
        assert r.status_code == 401
        r = client.get("/api/tunnels")
        assert r.status_code == 401

    def test_user_flow(self, client, test_user):
        headers = test_user["headers"]
        # 个人资料
        r = client.get("/api/user/profile", headers=headers)
        assert r.status_code == 200
        # 修改密码
        r = client.put("/api/user/password", headers=headers, json={"old_password": "testpass123", "new_password": "newpass456"})
        assert r.status_code == 200


class TestTunnels:
    """隧道 API。"""

    def test_create_list_detail(self, client, test_user):
        headers = test_user["headers"]
        # 获取节点
        r = client.get("/api/nodes", headers=headers)
        nodes = r.json()["data"]
        assert len(nodes) > 0
        node_id = nodes[0]["id"]

        # 创建隧道
        r = client.post("/api/tunnels", headers=headers, json={
            "name": "apitest-tcp",
            "node_id": node_id,
            "type": "tcp",
            "local_ip": "127.0.0.1",
            "local_port": 8080,
        })
        assert r.status_code == 201
        tunnel_id = r.json()["data"]["id"]

        # 列表
        r = client.get("/api/tunnels", headers=headers)
        assert any(t["id"] == tunnel_id for t in r.json()["data"])

        # 详情
        r = client.get(f"/api/tunnels/{tunnel_id}", headers=headers)
        assert r.status_code == 200

        # 生成配置
        r = client.post(f"/api/tunnels/{tunnel_id}/config", headers=headers)
        assert r.status_code == 200
        assert "serverAddr" in r.json()["data"]["config"]

        # 启动/停止
        r = client.post(f"/api/tunnels/{tunnel_id}/start", headers=headers)
        assert r.status_code == 200
        r = client.post(f"/api/tunnels/{tunnel_id}/stop", headers=headers)
        assert r.status_code == 200

        # 删除
        r = client.delete(f"/api/tunnels/{tunnel_id}", headers=headers)
        assert r.status_code == 204

    def test_cross_user_access_forbidden(self, client, test_user):
        """越权：A 用户不能访问 B 用户隧道。"""
        # 创建第一个用户的隧道
        headers_a = test_user["headers"]
        r = client.get("/api/nodes", headers=headers_a)
        node_id = r.json()["data"][0]["id"]
        r = client.post("/api/tunnels", headers=headers_a, json={
            "name": "tunnel-a", "node_id": node_id, "type": "tcp",
            "local_ip": "127.0.0.1", "local_port": 8081,
        })
        tunnel_id = r.json()["data"]["id"]

        # 第二个用户访问该隧道应 404（ORM 层 user_id 过滤防越权）
        suffix = random.randint(20000, 99999)
        client.post("/api/auth/register", json={"username": f"b{suffix}", "email": f"b{suffix}@test.com", "password": "testpass123"})
        from app.core.database import SessionLocal
        from app.models import EmailCode
        db = SessionLocal()
        code_row = db.query(EmailCode).filter(EmailCode.email == f"b{suffix}@test.com").order_by(EmailCode.id.desc()).first()
        code = code_row.code if code_row else "000000"
        db.close()
        client.post("/api/auth/email-verify", json={"email": f"b{suffix}@test.com", "code": code, "purpose": "register"})
        r = client.post("/api/auth/login", json={"username": f"b{suffix}", "password": "testpass123"})
        headers_b = {"Authorization": f"Bearer {r.json()['data']['token']}"}
        r = client.get(f"/api/tunnels/{tunnel_id}", headers=headers_b)
        assert r.status_code == 404


class TestSigninPoints:
    """签到与积分 API。"""

    def test_signin_and_points(self, client, test_user):
        headers = test_user["headers"]
        r = client.post("/api/signin", headers=headers)
        assert r.status_code == 200
        data = r.json()["data"]
        assert data["points"] == 10
        # 重复签到
        r = client.post("/api/signin", headers=headers)
        assert r.json()["business_code"] == 4001
        # 积分流水
        r = client.get("/api/points/logs", headers=headers)
        assert r.json()["data"]["total"] >= 1
        # 兑换规则
        r = client.get("/api/points/rules", headers=headers)
        assert r.json()["data"]["exchange_points"] == 300

    def test_exchange_insufficient(self, client, test_user):
        headers = test_user["headers"]
        # 新用户 0 积分，兑换应失败
        r = client.post("/api/points/exchange", headers=headers)
        assert r.json()["business_code"] == 4002


class TestAnnouncementsTickets:
    """公告与工单 API。"""

    def test_announcements_public(self, client):
        r = client.get("/api/announcements")
        assert r.status_code == 200

    def test_ticket_flow(self, client, test_user):
        headers = test_user["headers"]
        # 创建工单
        r = client.post("/api/tickets", headers=headers, json={"title": "测试工单", "content": "这是工单内容"})
        assert r.status_code == 201
        ticket_id = r.json()["data"]["id"]
        # 列表
        r = client.get("/api/tickets", headers=headers)
        assert any(t["id"] == ticket_id for t in r.json()["data"]["items"])
        # 回复
        r = client.post(f"/api/tickets/{ticket_id}/reply", headers=headers, json={"content": "补充信息"})
        assert r.status_code == 200
        # 关闭
        r = client.post(f"/api/tickets/{ticket_id}/close", headers=headers)
        assert r.status_code == 200
        # 关闭后回复失败
        r = client.post(f"/api/tickets/{ticket_id}/reply", headers=headers, json={"content": "x"})
        assert r.json()["business_code"] == 5001


class TestAdmin:
    """管理后台 API。"""

    def test_admin_required(self, client, test_user):
        """普通用户访问管理接口应 403。"""
        r = client.get("/api/admin/users", headers=test_user["headers"])
        assert r.status_code == 403

    def test_admin_dashboard(self, client, admin_headers):
        r = client.get("/api/admin/dashboard", headers=admin_headers)
        assert r.status_code == 200
        data = r.json()["data"]
        assert "summary" in data
        assert "traffic_series" in data

    def test_admin_users(self, client, admin_headers):
        r = client.get("/api/admin/users", headers=admin_headers)
        assert r.status_code == 200
        assert "items" in r.json()["data"]

    def test_admin_announcement(self, client, admin_headers):
        r = client.post("/api/admin/announcements", headers=admin_headers, json={
            "title": "测试公告",
            "content": "这是一条测试公告内容",
            "author": "技术部",
        })
        assert r.status_code == 201
        assert r.json()["data"]["author"] == "技术部"
        # 下线
        ann_id = r.json()["data"]["id"]
        r = client.post(f"/api/admin/announcements/{ann_id}/offline", headers=admin_headers)
        assert r.status_code == 200
        # 用户侧已下线公告不可见
        r = client.get(f"/api/announcements/{ann_id}")
        assert r.json()["business_code"] == 5002

    def test_admin_nodes(self, client, admin_headers):
        r = client.get("/api/admin/nodes", headers=admin_headers)
        assert r.status_code == 200
        assert len(r.json()["data"]) >= 1
        node_id = r.json()["data"][0]["id"]
        # 修改限速
        r = client.put(f"/api/admin/nodes/{node_id}/speed", headers=admin_headers, json={"speed_limit_mbps": 200})
        assert r.status_code == 200
        # 操作日志
        r = client.get("/api/admin/logs/operation", headers=admin_headers)
        assert r.status_code == 200

    def test_admin_config(self, client, admin_headers):
        r = client.put("/api/admin/config", headers=admin_headers, json={"key": "exchange_points", "value": "350"})
        assert r.status_code == 200
        r = client.get("/api/admin/config", headers=admin_headers)
        assert r.json()["data"]["exchange_points"] == "350"


class TestAgent:
    """内核联动 API。"""

    def test_agent_register_and_pull(self, client, admin_headers):
        # 获取节点 Agent Token
        r = client.get("/api/admin/nodes", headers=admin_headers)
        node = r.json()["data"][0]
        agent_token = node["agent_token"]

        # 注册
        r = client.post("/api/agent/register", json={
            "agent_token": agent_token,
            "name": node["name"],
            "address": "127.0.0.1",
            "port": 7000,
        })
        assert r.status_code == 200
        assert r.json()["data"]["node_id"] == node["id"]

        # 心跳
        r = client.post("/api/agent/heartbeat", headers={"Authorization": f"Bearer {agent_token}"}, json={
            "tunnels": [{"tunnel_id": 1, "online": True, "connections": 2, "in_delta": 100, "out_delta": 200}],
        })
        assert r.status_code == 200

        # 拉取配置
        r = client.get("/api/agent/tunnels", headers={"Authorization": f"Bearer {agent_token}"})
        assert r.status_code == 200
        assert "tunnels" in r.json()["data"]

    def test_agent_invalid_token(self, client):
        r = client.get("/api/agent/tunnels", headers={"Authorization": "Bearer agent_wrong"})
        assert r.status_code == 401


class TestClientAPI:
    """客户端 API。"""

    def test_client_login_and_tunnels(self, client, test_user):
        r = client.post("/api/client/login", json={"username": test_user["username"], "password": test_user["password"]})
        assert r.status_code == 200
        token = r.json()["data"]["token"]
        headers = {"Authorization": f"Bearer {token}"}
        r = client.get("/api/client/tunnels", headers=headers)
        assert r.status_code == 200
