"""asyncio 后台任务。

- 会员过期检查（每小时）：到期自动降级回免费版
- 流量日汇总（每 5 分钟）：Redis 增量写入 traffic_stats
- 节点心跳超时标记（每 60s 扫描）：超过 60s 无心跳标记离线
- 邮件队列消费（MailQueue 自带 worker）
"""
from __future__ import annotations

import asyncio
import logging
from datetime import date, datetime

from ..core.config import get_settings
from ..core.database import SessionLocal
from ..core.redis_client import redis_client
from ..models import Node, Plan, TrafficStat, User, UserPlanLog

logger = logging.getLogger("weavenet.tasks")
settings = get_settings()

# 心跳超时（秒），默认 60
HEARTBEAT_TIMEOUT = settings.node_heartbeat_timeout


async def check_plan_expiry() -> None:
    """会员到期检查：到期自动降级回免费版。"""
    try:
        with SessionLocal() as db:
            free_plan = db.query(Plan).filter(Plan.is_default.is_(True)).first()
            if free_plan is None:
                free_plan = db.query(Plan).order_by(Plan.sort).first()
            now = datetime.now()
            expired_users = (
                db.query(User)
                .filter(User.plan_expires_at.isnot(None), User.plan_expires_at < now)
                .all()
            )
            for user in expired_users:
                if free_plan is not None and user.plan_id != free_plan.id:
                    db.add(
                        UserPlanLog(
                            user_id=user.id,
                            plan_id=free_plan.id,
                            plan_name=free_plan.name,
                            reason="会员到期自动降级",
                        )
                    )
                    user.plan_id = free_plan.id
                    user.plan_expires_at = None
                    logger.info("用户 %s 会员到期，自动降级为 %s", user.username, free_plan.name)
            if expired_users:
                db.commit()
    except Exception as exc:  # noqa: BLE001
        logger.exception("会员到期检查异常: %s", exc)


async def aggregate_daily_traffic() -> None:
    """流量日汇总：读取 Redis 当日增量写入 traffic_stats。"""
    try:
        keys = redis_client.flush_all_today_traffic_keys()
        if not keys:
            return
        today = date.today().strftime("%Y-%m-%d")
        with SessionLocal() as db:
            for key in keys:
                try:
                    tunnel_id = int(key.rsplit(":", 1)[-1])
                except (ValueError, IndexError):
                    continue
                data = redis_client.get_and_clear_traffic(key)
                if not data["in"] and not data["out"]:
                    continue
                row = (
                    db.query(TrafficStat)
                    .filter(TrafficStat.tunnel_id == tunnel_id, TrafficStat.date == today)
                    .first()
                )
                if row is None:
                    db.add(
                        TrafficStat(
                            tunnel_id=tunnel_id,
                            date=today,
                            in_bytes=data["in"],
                            out_bytes=data["out"],
                        )
                    )
                else:
                    row.in_bytes += data["in"]
                    row.out_bytes += data["out"]
            db.commit()
    except Exception as exc:  # noqa: BLE001
        logger.exception("流量日汇总异常: %s", exc)


async def mark_offline_nodes() -> None:
    """节点心跳超时扫描：超过 60s 无心跳标记离线。"""
    try:
        with SessionLocal() as db:
            # 方案一：Redis 在线标记已过期
            stale = []
            nodes = db.query(Node).filter(Node.status.in_(["online", "maintenance"])).all()
            for node in nodes:
                if node.status == "online" and not redis_client.is_node_online(node.id):
                    stale.append(node)
            for node in stale:
                node.status = "offline"
                logger.warning("节点 %s 心跳超时，标记离线", node.name)
            if stale:
                db.commit()
            # 方案二：数据库兜底（Redis 降级时）
            from datetime import timedelta

            threshold = datetime.now() - timedelta(seconds=HEARTBEAT_TIMEOUT)
            if redis_client.degraded:
                stale_db = (
                    db.query(Node)
                    .filter(
                        Node.status == "online",
                        Node.last_heartbeat_at.isnot(None),
                        Node.last_heartbeat_at < threshold,
                    )
                    .all()
                )
                for node in stale_db:
                    node.status = "offline"
                    logger.warning("节点 %s 数据库心跳超时，标记离线", node.name)
                if stale_db:
                    db.commit()
    except Exception as exc:  # noqa: BLE001
        logger.exception("节点心跳扫描异常: %s", exc)


async def scheduler_loop() -> None:
    """后台任务调度主循环。"""
    await asyncio.gather(
        _run_every("plan-expiry", check_plan_expiry, 3600),
        _run_every("traffic-aggregate", aggregate_daily_traffic, 300),
        _run_every("node-heartbeat-scan", mark_offline_nodes, 60),
    )


async def _run_every(name: str, coro, interval: int) -> None:
    """按固定间隔循环执行任务，首次先执行一次。"""
    while True:
        try:
            await coro()
        except asyncio.CancelledError:
            raise
        except Exception as exc:  # noqa: BLE001
            logger.exception("任务 %s 执行异常: %s", name, exc)
        await asyncio.sleep(interval)
