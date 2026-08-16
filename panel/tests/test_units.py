"""单元测试：签到规则、套餐校验、Token 生成、限速计算、frpc 配置。"""
from __future__ import annotations

import sys
from datetime import date, timedelta
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))


# ---------- Token 生成 ----------

class TestToken:
    def test_session_token_length(self):
        from app.core.security import generate_token

        token = generate_token(64)
        assert len(token) == 64
        assert token.isalnum()

    def test_token_unique(self):
        from app.core.security import generate_token

        tokens = {generate_token(64) for _ in range(100)}
        assert len(tokens) == 100

    def test_agent_token_prefix(self):
        from app.core.security import generate_agent_token

        assert generate_agent_token().startswith("agent_")

    def test_email_code_numeric(self):
        from app.core.security import generate_email_code

        code = generate_email_code()
        assert code.isdigit()
        assert len(code) == 6

    def test_emergency_key_length(self):
        from app.core.security import generate_emergency_key

        assert len(generate_emergency_key()) >= 32


# ---------- 密码哈希 ----------

class TestPassword:
    def test_hash_and_verify(self):
        from app.core.security import hash_password, verify_password

        hashed = hash_password("testpass123")
        assert hashed != "testpass123"
        assert verify_password("testpass123", hashed)
        assert not verify_password("wrongpass", hashed)

    def test_bcrypt_rounds(self):
        import bcrypt

        from app.core.security import hash_password

        hashed = hash_password("testpass123")
        assert bcrypt.checkpw(b"testpass123", hashed.encode())


# ---------- 签到规则 ----------

class TestSignin:
    def test_first_signin(self):
        from app.core.database import SessionLocal
        from app.models import Plan, User
        from app.services.points import can_signin, do_signin
        from app.core.security import hash_password

        with SessionLocal() as db:
            free_plan = db.query(Plan).filter(Plan.is_default.is_(True)).first()
            user = User(
                username=f"signin_test_{id(db)}",
                email=f"signin_test_{id(db)}@test.com",
                password_hash=hash_password("testpass123"),
                plan_id=free_plan.id if free_plan else 1,
            )
            db.add(user)
            db.commit()
            db.refresh(user)
            # 首次签到
            status = can_signin(db, user)
            assert status["today_signed"] is False
            result = do_signin(db, user)
            assert result["points"] == 10
            assert result["continuous_days"] == 1
            assert result["bonus"] == 0
            # 重复签到应抛业务错误
            with pytest.raises(Exception):
                do_signin(db, user)
            # 状态更新
            status2 = can_signin(db, user)
            assert status2["today_signed"] is True
            db.delete(user)
            db.commit()


# ---------- 套餐额度 ----------

class TestQuota:
    def test_plan_limits(self):
        from app.core.database import SessionLocal
        from app.models import Plan

        with SessionLocal() as db:
            plans = {p.name: p for p in db.query(Plan).all()}
            free = plans["免费版"]
            assert free.tunnel_limit == 3
            assert free.domain_limit == 1
            assert free.speed_limit_mbps == 8
            normal = plans["普通会员"]
            assert normal.tunnel_limit == 6
            assert normal.domain_limit == 4
            assert normal.speed_limit_mbps == 16
            super_v = plans["超级会员"]
            assert super_v.tunnel_limit == 14
            assert super_v.domain_limit == 16
            assert super_v.speed_limit_mbps == 32


# ---------- 限速换算 ----------

class TestBandwidth:
    def test_kbps_conversion(self):
        from app.services.frp_config import tunnel_bandwidth
        from app.models import Plan, User

        plan = Plan(id=1, name="免费版", speed_limit_mbps=8)
        user = User(id=1, plan=plan)
        assert tunnel_bandwidth(user) == "8000KB"


# ---------- frpc 配置生成 ----------

class TestFrpcConfig:
    def test_generate_config(self):
        from app.models import Node, Plan, Tunnel, User
        from app.services.frp_config import generate_frpc_config

        plan = Plan(id=1, name="免费版", speed_limit_mbps=8)
        user = User(id=1, username="test", plan=plan, password_hash="x" * 60)
        node = Node(id=1, name="节点", address="127.0.0.1", port=7000)
        tunnel = Tunnel(
            id=1,
            name="web",
            type="tcp",
            local_ip="127.0.0.1",
            local_port=8080,
            remote_port=20000,
            kcp=False,
            encryption=True,
            compression=False,
            node=node,
            user=user,
        )
        cfg = generate_frpc_config(tunnel, user, node)
        assert "serverAddr" in cfg
        assert "20000" in cfg
        assert "token" in cfg
