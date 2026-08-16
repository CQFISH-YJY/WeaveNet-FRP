"""安全工具：密码哈希、会话 Token、随机码。"""
from __future__ import annotations

import secrets
import string

import bcrypt

from .config import get_settings

settings = get_settings()


def hash_password(password: str) -> str:
    """bcrypt 哈希密码（不少于 10 轮）。"""
    rounds = max(10, settings.bcrypt_rounds)
    salt = bcrypt.gensalt(rounds=rounds)
    return bcrypt.hashpw(password.encode("utf-8"), salt).decode("utf-8")


def verify_password(password: str, password_hash: str) -> bool:
    """校验密码。"""
    try:
        return bcrypt.checkpw(password.encode("utf-8"), password_hash.encode("utf-8"))
    except (ValueError, TypeError):
        return False


def generate_token(length: int = 64) -> str:
    """生成 64 位随机会话 Token（字母数字）。"""
    alphabet = string.ascii_letters + string.digits
    return "".join(secrets.choice(alphabet) for _ in range(length))


def generate_agent_token() -> str:
    """生成节点 Agent Token。"""
    return "agent_" + generate_token(48)


def generate_email_code(length: int = 6) -> str:
    """生成数字验证码。"""
    return "".join(secrets.choice(string.digits) for _ in range(length))


def generate_emergency_key(length: int = 48) -> str:
    """生成应急服务预共享密钥（>=32 位强随机）。"""
    alphabet = string.ascii_letters + string.digits + "!@#$%^&*"
    return "".join(secrets.choice(alphabet) for _ in range(length))


def generate_secret_key(length: int = 12) -> str:
    """生成 stcp/xtcp 访问密钥。"""
    alphabet = string.ascii_letters + string.digits
    return "".join(secrets.choice(alphabet) for _ in range(length))
