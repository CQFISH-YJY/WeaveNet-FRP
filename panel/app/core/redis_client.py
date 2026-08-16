"""Redis 封装。

Redis 承担会话、缓存、实时状态三类职责。当 Redis 不可用时自动降级
为内存实现，保证面板核心功能不中断。
"""
from __future__ import annotations

import json
import time
from collections import OrderedDict
from datetime import datetime, timedelta
from typing import Any

import redis as redis_lib

from .config import get_settings

settings = get_settings()


class MemoryStore:
    """Redis 不可用时的内存降级实现（带简单 LRU 清理）。"""

    def __init__(self, max_items: int = 20000) -> None:
        self._store: OrderedDict[str, tuple[float, Any]] = OrderedDict()
        self._max_items = max_items

    def _evict(self) -> None:
        while len(self._store) > self._max_items:
            self._store.popitem(last=False)

    def setex(self, key: str, seconds: int, value: Any) -> None:
        self._store[key] = (time.time() + seconds, value)
        self._evict()

    def get(self, key: str) -> Any | None:
        item = self._store.get(key)
        if item is None:
            return None
        expire_at, value = item
        if expire_at < time.time():
            self._store.pop(key, None)
            return None
        self._store.move_to_end(key)
        return value

    def delete(self, *keys: str) -> None:
        for key in keys:
            self._store.pop(key, None)

    def incr(self, key: str) -> int:
        item = self._store.get(key)
        if item is None:
            self._store[key] = (time.time() + 86400, 1)
            return 1
        expire_at, value = item
        self._store[key] = (expire_at, value + 1)
        self._store.move_to_end(key)
        return value + 1

    def expire(self, key: str, seconds: int) -> None:
        item = self._store.get(key)
        if item is not None:
            self._store[key] = (time.time() + seconds, item[1])

    def keys(self, pattern: str = "*") -> list[str]:
        return [k for k in self._store if _glob_match(pattern, k)]


def _glob_match(pattern: str, value: str) -> bool:
    """极简 glob 匹配（支持 * 与 ?）。"""
    import fnmatch

    return fnmatch.fnmatchcase(value, pattern)


class RedisClient:
    """Redis 客户端，提供会话/缓存/实时状态操作，带降级能力。"""

    def __init__(self) -> None:
        self._client: redis_lib.Redis | None = None
        self._memory = MemoryStore()
        self._degraded = False
        self._connect()

    def _connect(self) -> None:
        try:
            self._client = redis_lib.Redis.from_url(
                settings.redis_url, decode_responses=True, socket_timeout=2
            )
            self._client.ping()
            self._degraded = False
        except Exception:
            self._client = None
            self._degraded = True

    @property
    def degraded(self) -> bool:
        """当前是否运行在内存降级模式。"""
        return self._degraded

    def reconnect(self) -> None:
        self._connect()

    # ---------- 基础操作 ----------

    def setex(self, key: str, seconds: int, value: Any) -> None:
        if self._client is not None:
            try:
                self._client.setex(key, seconds, value)
                return
            except Exception:
                self._connect()
        self._memory.setex(key, seconds, value)

    def get(self, key: str) -> Any | None:
        if self._client is not None:
            try:
                return self._client.get(key)
            except Exception:
                self._connect()
        return self._memory.get(key)

    def delete(self, *keys: str) -> None:
        if self._client is not None:
            try:
                self._client.delete(*keys)
                return
            except Exception:
                self._connect()
        self._memory.delete(*keys)

    def incr(self, key: str) -> int:
        if self._client is not None:
            try:
                return int(self._client.incr(key))
            except Exception:
                self._connect()
        return self._memory.incr(key)

    def expire(self, key: str, seconds: int) -> None:
        if self._client is not None:
            try:
                self._client.expire(key, seconds)
                return
            except Exception:
                self._connect()
        self._memory.expire(key, seconds)

    # ---------- JSON 便捷 ----------

    def set_json(self, key: str, seconds: int, data: Any) -> None:
        self.setex(key, seconds, json.dumps(data, ensure_ascii=False, default=str))

    def get_json(self, key: str) -> Any | None:
        raw = self.get(key)
        if raw is None:
            return None
        try:
            return json.loads(raw)
        except (TypeError, ValueError):
            return None

    # ---------- 实时状态 ----------

    def set_node_online(self, node_id: int) -> None:
        """刷新节点在线状态（TTL 为心跳超时的 3 倍）。"""
        ttl = max(60, settings.node_heartbeat_timeout * 3)
        self.setex(f"node:online:{node_id}", ttl, "1")

    def is_node_online(self, node_id: int) -> bool:
        return self.get(f"node:online:{node_id}") is not None

    def set_tunnel_runtime(self, tunnel_id: int, data: dict[str, Any], ttl: int = 90) -> None:
        """缓存隧道实时运行数据（连接数、流量等）。"""
        self.set_json(f"tunnel:runtime:{tunnel_id}", ttl, data)

    def get_tunnel_runtime(self, tunnel_id: int) -> dict[str, Any] | None:
        return self.get_json(f"tunnel:runtime:{tunnel_id}")

    def incr_traffic(self, tunnel_id: int, in_delta: int, out_delta: int) -> None:
        """累加当日流量增量。"""
        today = datetime.now().strftime("%Y-%m-%d")
        key = f"traffic:today:{today}:{tunnel_id}"
        # 使用 pipeline 原子累加
        if self._client is not None:
            try:
                pipe = self._client.pipeline()
                pipe.hincrby(key, "in", in_delta)
                pipe.hincrby(key, "out", out_delta)
                pipe.expire(key, 86400 * 2)
                pipe.execute()
                return
            except Exception:
                self._connect()
        raw = self._memory.get(key)
        current = raw if isinstance(raw, dict) else {}
        self._memory.setex(
            key,
            86400 * 2,
            {
                "in": current.get("in", 0) + in_delta,
                "out": current.get("out", 0) + out_delta,
            },
        )

    def get_today_traffic(self, tunnel_id: int) -> dict[str, int]:
        today = datetime.now().strftime("%Y-%m-%d")
        key = f"traffic:today:{today}:{tunnel_id}"
        raw = self.get_json(key)
        if raw is None:
            return {"in": 0, "out": 0}
        return {"in": int(raw.get("in", 0)), "out": int(raw.get("out", 0))}

    def flush_all_today_traffic_keys(self) -> list[str]:
        """返回所有当日流量 key（供日汇总任务消费）。"""
        today = datetime.now().strftime("%Y-%m-%d")
        pattern = f"traffic:today:{today}:*"
        if self._client is not None:
            try:
                return list(self._client.keys(pattern))
            except Exception:
                self._connect()
        return self._memory.keys(pattern)

    def get_and_clear_traffic(self, key: str) -> dict[str, int]:
        """读取并清除单条流量缓存（用于入库聚合）。"""
        raw = self.get_json(key)
        self.delete(key)
        if raw is None:
            return {"in": 0, "out": 0}
        return {"in": int(raw.get("in", 0)), "out": int(raw.get("out", 0))}

    # ---------- 会话 ----------

    def set_session(self, token: str, user_id: int, days: int | None = None) -> None:
        ttl = timedelta(days=days or settings.session_days).total_seconds()
        self.set_json(f"session:{token}", int(ttl), {"user_id": user_id})

    def get_session_user(self, token: str) -> int | None:
        data = self.get_json(f"session:{token}")
        if data is None:
            return None
        return int(data.get("user_id", 0))

    def del_session(self, token: str) -> None:
        self.delete(f"session:{token}")

    # ---------- 限流 ----------

    def rate_limit_hit(self, bucket: str, limit: int, window_seconds: int = 60) -> bool:
        """滑动窗口限流：达到 limit 则返回 True 表示触发限流。"""
        key = f"ratelimit:{bucket}"
        count = self.incr(key)
        if count == 1:
            self.expire(key, window_seconds)
        return count > limit


redis_client = RedisClient()
