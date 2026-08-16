"""WeaveNet 织网穿透 应急服务（逃生通道）。

独立于主面板的轻量应急进程，零第三方依赖（仅标准库），
面板 / 数据库 / Redis 全部故障时仍可响应。

功能：
  GET  /health           存活检查
  GET  /status           系统状态（CPU/内存/磁盘/容器状态）
  GET  /logs             拉取服务日志
  GET  /data             直连 SQLite 拉取只读数据
  POST /restart          重启服务
  POST /stop             停止服务
  POST /start            启动服务
  POST /reboot           重启服务器
  POST /exec             受限白名单命令执行

安全：
  强随机预共享密钥（>=32 位）鉴权，连续失败 5 次锁定 10 分钟，
  危险操作（reboot/exec）需二次确认参数。

运行：python emergency.py --port 9001
文档：CQFISH&喵酱出品
"""
from __future__ import annotations

import argparse
import base64
import hashlib
import hmac
import json
import os
import platform
import shutil
import sqlite3
import subprocess
import sys
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

# ---------- 配置 ----------

BASE_DIR = Path(__file__).resolve().parent
CONFIG_FILE = BASE_DIR / "emergency.conf"
LOG_FILE = BASE_DIR / "emergency.log"

DEFAULT_CONFIG = {
    "port": 9001,
    "secret_key": "",  # 首次启动自动生成
    "ip_whitelist": "",  # 逗号分隔，留空不限制
    "db_path": "",  # SQLite 数据库路径（用于 /data）
    "service_command": {},  # JSON: 服务名 -> 启动命令列表
}

MAX_FAILS = 5
LOCK_SECONDS = 600

# ---------- 配置加载 ----------


def load_config() -> dict:
    """加载配置，不存在则生成默认并写入。"""
    cfg = dict(DEFAULT_CONFIG)
    if CONFIG_FILE.exists():
        try:
            for line in CONFIG_FILE.read_text(encoding="utf-8").splitlines():
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, _, value = line.partition("=")
                cfg[key.strip()] = value.strip()
        except OSError:
            pass
    # 首次启动生成强随机密钥
    if not cfg["secret_key"] or len(cfg["secret_key"]) < 32:
        cfg["secret_key"] = base64.urlsafe_b64encode(os.urandom(48)).decode()[:48]
        save_config(cfg)
    return cfg


def save_config(cfg: dict) -> None:
    """写回配置。"""
    lines = [
        "# WeaveNet 应急服务配置",
        f"port = {cfg.get('port', 9001)}",
        f"secret_key = {cfg.get('secret_key', '')}",
        f"ip_whitelist = {cfg.get('ip_whitelist', '')}",
        f"db_path = {cfg.get('db_path', '')}",
        f"service_command = {cfg.get('service_command', '{}')}",
        "",
    ]
    CONFIG_FILE.write_text("\n".join(lines), encoding="utf-8")


# ---------- 审计日志 ----------


def audit(action: str, detail: str = "") -> None:
    """记录审计日志（本地文件）。"""
    ts = time.strftime("%Y-%m-%d %H:%M:%S")
    line = f"{ts} [{action}] {detail}\n"
    try:
        with open(LOG_FILE, "a", encoding="utf-8") as f:
            f.write(line)
    except OSError:
        pass


# ---------- 鉴权 ----------


class AuthGuard:
    """鉴权与锁定机制。"""

    def __init__(self, secret_key: str) -> None:
        self.secret_key = secret_key.encode()
        self.lock = threading.Lock()
        self.fails = 0
        self.locked_until = 0.0

    def verify(self, request_secret: str) -> bool:
        """校验请求密钥，处理连续失败锁定。"""
        now = time.time()
        with self.lock:
            if now < self.locked_until:
                return False
            if self.fails >= MAX_FAILS:
                self.locked_until = now + LOCK_SECONDS
                self.fails = 0
                audit("lock", f"连续失败 {MAX_FAILS} 次，锁定 {LOCK_SECONDS} 秒")
                return False
        ok = hmac.compare_digest(request_secret.encode(), self.secret_key)
        if not ok:
            with self.lock:
                self.fails += 1
        else:
            with self.lock:
                self.fails = 0
        return ok

    @property
    def locked(self) -> bool:
        return time.time() < self.locked_until


# ---------- 系统状态 ----------


def system_status() -> dict:
    """采集系统状态。"""
    status = {
        "time": time.strftime("%Y-%m-%d %H:%M:%S"),
        "platform": platform.platform(),
        "python": platform.python_version(),
        "cpu_count": os.cpu_count() or 0,
        "loadavg": None,
        "memory": {},
        "disk": {},
        "containers": {},
    }
    try:
        if hasattr(os, "getloadavg"):
            status["loadavg"] = list(os.getloadavg())
    except OSError:
        pass
    # 内存（Windows/Linux 差异处理）
    try:
        if sys.platform.startswith("linux"):
            with open("/proc/meminfo", encoding="utf-8") as f:
                for line in f:
                    parts = line.split(":")
                    if len(parts) == 2:
                        key = parts[0].strip()
                        val = parts[1].strip().split()[0]
                        status["memory"][key] = int(val)
                    if len(status["memory"]) >= 4:
                        break
        else:
            import ctypes

            class MEMORYSTATUS(ctypes.Structure):  # noqa: N801
                _fields_ = [
                    ("dwLength", ctypes.c_ulong),
                    ("dwMemoryLoad", ctypes.c_ulong),
                    ("ullTotalPhys", ctypes.c_ulonglong),
                    ("ullAvailPhys", ctypes.c_ulonglong),
                    ("ullTotalPageFile", ctypes.c_ulonglong),
                    ("ullAvailPageFile", ctypes.c_ulonglong),
                    ("ullTotalVirtual", ctypes.c_ulonglong),
                    ("ullAvailVirtual", ctypes.c_ulonglong),
                    ("ullAvailExtendedVirtual", ctypes.c_ulonglong),
                ]

            ms = MEMORYSTATUS()
            ms.dwLength = ctypes.sizeof(MEMORYSTATUS)
            ctypes.windll.kernel32.GlobalMemoryStatusEx(ctypes.byref(ms))
            status["memory"] = {
                "dwMemoryLoad": ms.dwMemoryLoad,
                "ullTotalPhys": ms.ullTotalPhys,
                "ullAvailPhys": ms.ullAvailPhys,
            }
    except Exception as exc:  # noqa: BLE001
        status["memory"] = {"error": str(exc)}

    # 磁盘
    try:
        total, used, free = shutil.disk_usage(str(BASE_DIR))
        status["disk"] = {
            "total": total,
            "used": used,
            "free": free,
        }
    except OSError:
        pass
    # 容器状态
    if shutil.which("docker"):
        for name in ("weave-panel", "weave-redis", "weave-frps", "weave-emergency"):
            try:
                result = subprocess.run(
                    ["docker", "inspect", "-f", "{{.State.Status}}", name],
                    capture_output=True, text=True, timeout=5,
                )
                status["containers"][name] = result.stdout.strip() or "not found"
            except (subprocess.SubprocessError, OSError):
                status["containers"][name] = "unknown"
    return status


# ---------- HTTP 服务 ----------


class EmergencyHandler(BaseHTTPRequestHandler):
    """应急服务 HTTP 处理器。"""

    guard: AuthGuard = None
    cfg: dict = {}

    # ---- 工具方法 ----

    def _send(self, code: int, payload: dict) -> None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _ok(self, data=None, message: str = "ok") -> None:
        self._send(200, {"code": 0, "message": message, "data": data})

    def _fail(self, code: int, message: str) -> None:
        self._send(code, {"code": code, "message": message, "data": None})

    def _auth(self) -> bool:
        """校验密钥与 IP 白名单。"""
        # IP 白名单
        whitelist = self.cfg.get("ip_whitelist", "").strip()
        if whitelist:
            allowed = [ip.strip() for ip in whitelist.split(",") if ip.strip()]
            client_ip = self.client_address[0]
            if allowed and client_ip not in allowed:
                audit("deny_ip", f"{client_ip} 不在白名单")
                self._fail(403, "IP 不在白名单内")
                return False
        secret = self.headers.get("X-Emergency-Key", "")
        if not self.guard.verify(secret):
            if self.guard.locked:
                self._fail(423, "已锁定，请稍后再试")
            else:
                self._fail(401, "密钥错误")
            return False
        return True

    def _query(self, name: str, default: str = "") -> str:
        from urllib.parse import parse_qs, urlparse

        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)
        values = params.get(name, [])
        return values[0] if values else default

    def _confirm_token(self, action: str) -> bool:
        """危险操作二次确认参数。"""
        token = self._query("confirm")
        expected = hashlib.sha256(f"weavenet:{action}".encode()).hexdigest()[:16]
        if token != expected:
            self._fail(400, f"危险操作需二次确认，请携带 confirm={expected}")
            return False
        return True

    # ---- GET 接口 ----

    def do_GET(self) -> None:
        try:
            self._route_get()
        except (BrokenPipeError, ConnectionResetError):
            pass
        except Exception as exc:  # noqa: BLE001
            audit("error", f"GET {self.path}: {exc}")
            self._fail(500, "内部错误")

    def _route_get(self) -> None:
        path = self.path.split("?")[0]
        if path == "/health":
            self._ok({"status": "alive", "service": "emergency"}, "存活")
            return
        if not self._auth():
            return
        if path == "/status":
            audit("status", f"来自 {self.client_address[0]}")
            self._ok(system_status(), "系统状态")
        elif path == "/logs":
            service = self._query("service", "panel")
            lines = int(self._query("lines", "200") or 200)
            self._ok({"service": service, "log": self._tail_log(service, lines)}, "日志")
        elif path == "/data":
            self._read_data()
        else:
            self._fail(404, "接口不存在")

    def _tail_log(self, service: str, lines: int) -> str:
        """读取服务日志尾部。"""
        log_path = BASE_DIR / f"{service}.log"
        if not log_path.exists():
            return "(日志文件不存在)"
        try:
            data = log_path.read_text(encoding="utf-8", errors="ignore").splitlines()
            return "\n".join(data[-max(1, min(lines, 2000)) :])
        except OSError:
            return "(读取失败)"

    def _read_data(self) -> None:
        """直连 SQLite 拉取只读数据。"""
        data_type = self._query("type", "users")
        db_path = self.cfg.get("db_path") or ""
        if not db_path or not Path(db_path).exists():
            self._fail(400, "未配置数据库路径")
            return
        if data_type not in ("users", "tunnels", "nodes"):
            self._fail(400, "type 仅支持 users|tunnels|nodes")
            return
        table = {"users": "users", "tunnels": "tunnels", "nodes": "nodes"}[data_type]
        limit = min(int(self._query("limit", "100") or 100), 1000)
        try:
            conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True, timeout=5)
            cur = conn.cursor()
            cur.execute(f"SELECT * FROM {table} LIMIT ?", (limit,))
            cols = [d[0] for d in cur.description]
            rows = [dict(zip(cols, row)) for row in cur.fetchall()]
            conn.close()
            audit("data", f"读取 {table} {len(rows)} 行")
            self._ok({"type": data_type, "count": len(rows), "rows": rows}, "数据")
        except sqlite3.Error as exc:
            self._fail(500, f"数据库读取失败: {exc}")

    # ---- POST 接口 ----

    def do_POST(self) -> None:
        try:
            self._route_post()
        except (BrokenPipeError, ConnectionResetError):
            pass
        except Exception as exc:  # noqa: BLE001
            audit("error", f"POST {self.path}: {exc}")
            self._fail(500, "内部错误")

    def _route_post(self) -> None:
        path = self.path.split("?")[0]
        if not self._auth():
            return
        if path == "/restart":
            if not self._confirm_token("restart"):
                return
            service = self._query("service", "panel")
            audit("restart", f"{service} 重启（二次确认通过）")
            self._ok({"service": service}, "重启指令已执行")
            threading.Thread(target=self._restart_service, args=(service,), daemon=True).start()
        elif path == "/stop":
            service = self._query("service", "all")
            audit("stop", f"{service} 停止")
            self._ok({"service": service}, "停止指令已执行")
            threading.Thread(target=self._stop_service, args=(service,), daemon=True).start()
        elif path == "/start":
            service = self._query("service", "panel")
            audit("start", f"{service} 启动")
            self._ok({"service": service}, "启动指令已执行")
            threading.Thread(target=self._start_service, args=(service,), daemon=True).start()
        elif path == "/reboot":
            if not self._confirm_token("reboot"):
                return
            audit("reboot", "服务器重启（二次确认通过）")
            self._ok({}, "服务器重启指令已执行")
            threading.Thread(target=self._reboot, daemon=True).start()
        elif path == "/exec":
            if not self._confirm_token("exec"):
                return
            cmd = self._query("cmd", "")
            if not self._is_allowed_command(cmd):
                self._fail(400, "命令不在白名单内")
                return
            audit("exec", f"执行: {cmd}")
            threading.Thread(target=self._exec_cmd, args=(cmd,), daemon=True).start()
            self._ok({"cmd": cmd}, "命令已执行")
        else:
            self._fail(404, "接口不存在")

    # ---- 服务操作 ----

    def _service_command(self, service: str) -> list[str]:
        """解析服务的启停命令。"""
        try:
            cmds = json.loads(self.cfg.get("service_command", "{}") or "{}")
            return list(cmds.get(service, []))
        except (ValueError, TypeError):
            return []

    def _restart_service(self, service: str) -> None:
        cmds = self._service_command(service)
        if not cmds:
            audit("restart_skip", f"{service} 未配置命令，跳过")
            return
        self._run_commands(cmds)

    def _stop_service(self, service: str) -> None:
        cmds = self._service_command(service)
        if not cmds:
            audit("stop_skip", f"{service} 未配置命令，跳过")
            return
        self._run_commands(cmds)

    def _start_service(self, service: str) -> None:
        cmds = self._service_command(service)
        if not cmds:
            audit("start_skip", f"{service} 未配置命令，跳过")
            return
        self._run_commands(cmds)

    def _run_commands(self, cmds: list[str]) -> None:
        for cmd in cmds:
            try:
                subprocess.Popen(cmd, shell=True)
            except OSError as exc:
                audit("cmd_error", f"{cmd}: {exc}")

    def _is_allowed_command(self, cmd: str) -> bool:
        """受限白名单命令。"""
        allowed_prefixes = (
            "systemctl status", "docker ps", "docker logs", "free -m",
            "df -h", "uptime", "ps aux", "netstat -tlnp", "cat /etc/os-release",
        )
        return any(cmd.startswith(prefix) for prefix in allowed_prefixes)

    def _exec_cmd(self, cmd: str) -> None:
        try:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=30)
            audit("exec_result", f"exit={result.returncode} out={result.stdout[:500]} err={result.stderr[:200]}")
        except subprocess.SubprocessError as exc:
            audit("exec_error", str(exc))

    def _reboot(self) -> None:
        try:
            if sys.platform.startswith("linux"):
                subprocess.Popen(["shutdown", "-r", "now"])
            else:
                subprocess.Popen(["shutdown", "/r", "/t", "5"])
        except OSError as exc:
            audit("reboot_error", str(exc))

    def log_message(self, fmt, *args) -> None:  # noqa: ANN001
        """静默默认日志。"""


def main() -> None:
    """应急服务入口。"""
    parser = argparse.ArgumentParser(description="WeaveNet 应急服务")
    parser.add_argument("--port", type=int, default=0, help="监听端口（默认取配置）")
    args = parser.parse_args()

    cfg = load_config()
    if args.port:
        cfg["port"] = args.port

    guard = AuthGuard(cfg["secret_key"])
    EmergencyHandler.guard = guard
    EmergencyHandler.cfg = cfg

    port = int(cfg.get("port", 9001))
    server = ThreadingHTTPServer(("0.0.0.0", port), EmergencyHandler)
    audit("start", f"应急服务启动，端口 {port}")
    print(f"WeaveNet 应急服务运行于 http://0.0.0.0:{port}")
    print(f"密钥: {cfg['secret_key']}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        audit("stop", "应急服务停止")
        server.shutdown()


if __name__ == "__main__":
    main()
