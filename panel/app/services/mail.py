"""邮件发送服务。

SMTP 失败进入内存队列重试 3 次；开发环境未配置 SMTP 时，
将邮件内容打印到日志以便联调。
"""
from __future__ import annotations

import asyncio
import smtplib
import time
from email.mime.text import MIMEText
from email.utils import formataddr

from ..core.config import get_settings
from ..core.logger import get_logger

logger = get_logger("weavenet.mail")

settings = get_settings()

MAX_RETRIES = 3


class MailQueue:
    """内存邮件队列 + 异步发送。"""

    def __init__(self) -> None:
        self._queue: asyncio.Queue[dict] = asyncio.Queue(maxsize=1000)
        self._worker: asyncio.Task | None = None

    def start(self) -> None:
        if self._worker is None or self._worker.done():
            self._worker = asyncio.create_task(self._run())

    async def stop(self) -> None:
        if self._worker:
            self._worker.cancel()
            try:
                await self._worker
            except asyncio.CancelledError:
                pass
            self._worker = None

    async def _run(self) -> None:
        while True:
            item = await self._queue.get()
            to_email, subject, body = (
                item["to"],
                item["subject"],
                item["body"],
            )
            ok = await asyncio.to_thread(self._send_once, to_email, subject, body)
            retries = item.get("retries", 0)
            if not ok and retries < MAX_RETRIES:
                item["retries"] = retries + 1
                logger.warning("邮件发送失败，入队重试 %s/%s: %s", retries + 1, MAX_RETRIES, to_email)
                await asyncio.sleep(2)
                await self._queue.put(item)
            elif not ok:
                logger.error("邮件发送最终失败: %s", to_email)
            self._queue.task_done()

    def _send_once(self, to_email: str, subject: str, body: str) -> bool:
        if not settings.smtp_host or not settings.smtp_from:
            # 开发环境未配置 SMTP，打印到日志
            logger.info("[邮件占位] 收件人=%s 主题=%s 内容=%s", to_email, subject, body[:200])
            return True
        try:
            msg = MIMEText(body, "plain", "utf-8")
            msg["Subject"] = subject
            msg["From"] = formataddr((settings.app_name, settings.smtp_from))
            msg["To"] = to_email
            if settings.smtp_use_ssl:
                server = smtplib.SMTP_SSL(
                    settings.smtp_host, settings.smtp_port, timeout=15
                )
            else:
                server = smtplib.SMTP(
                    settings.smtp_host, settings.smtp_port, timeout=15
                )
                if settings.smtp_use_tls:
                    server.starttls()
            if settings.smtp_user:
                server.login(settings.smtp_user, settings.smtp_password)
            server.sendmail(settings.smtp_from, [to_email], msg.as_string())
            server.quit()
            return True
        except Exception as exc:
            logger.warning("SMTP 发送异常: %s", exc)
            return False

    def send(self, to_email: str, subject: str, body: str) -> None:
        """入队发送（非阻塞）。"""
        self.start()
        try:
            self._queue.put_nowait({"to": to_email, "subject": subject, "body": body, "retries": 0})
        except asyncio.QueueFull:
            logger.error("邮件队列已满，丢弃邮件: %s", to_email)


mail_queue = MailQueue()


def send_verification_code(to_email: str, code: str, purpose: str) -> None:
    """发送验证码邮件。"""
    purpose_text = {
        "register": "账号注册激活",
        "reset_password": "找回密码",
        "change_email": "修改邮箱",
    }.get(purpose, "账号验证")
    subject = f"【{settings.app_name}】{purpose_text}验证码"
    body = (
        f"您好：\n\n"
        f"您在 {settings.app_name} 的{purpose_text}验证码为：{code}\n"
        f"验证码自发送起 5 分钟内有效，请及时完成验证。\n\n"
        f"如非本人操作，请忽略本邮件。\n"
        f"—— {settings.app_name} 团队"
    )
    mail_queue.send(to_email, subject, body)
