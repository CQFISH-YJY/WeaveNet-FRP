'use strict';

// WeaveNet 织网穿透 桌面客户端 主进程
// 职责：创建窗口、处理 IPC、面板 API 网络请求、托管 frpc 子进程、状态轮询与断线重连

const { app, BrowserWindow, ipcMain, shell, dialog } = require('electron');
const path = require('path');
const fs = require('fs');
const os = require('os');
const http = require('http');
const https = require('https');
const { spawn } = require('child_process');

// ===================== 全局状态 =====================

let mainWindow = null;
// 内置面板服务器地址（分发版固定，不落盘、不可在界面修改，客户无感知）
const DEFAULT_SERVER_URL = 'http://64.90.4.13:8100';
const serverUrl = DEFAULT_SERVER_URL;
let customFrpcPath = '';                 // 用户自定义 frpc 路径（可空，自动查找）
let token = '';                          // 登录令牌，仅主进程持有，绝不下发渲染进程
let currentUser = null;                  // 当前登录用户信息

// 应用设置默认值（持久化到 state.json 的 settings 字段）
const DEFAULT_SETTINGS = {
  autoCheckUpdate: true,                 // 启动时自动检测更新
  minimizeToTray: true,                  // 关闭窗口最小化到托盘
  processGuard: true,                    // 进程守护：隧道意外退出自动重启
  frpcLogLevel: 'info',                  // frpc 日志输出等级
  autoRestartOnChange: true,             // 修改隧道后自动重启映射
  theme: 'system',                       // system | light | dark
  sidebarStyle: 'default',               // default | compact
  showTitleBar: true,                    // 显示顶部标题栏
  soundEnabled: true,                    // 操作提示音效
  backgroundType: 'none',                // none | image | video | folder
  backgroundValue: '',                   // 背景路径
  bypassProxy: false,                    // 绕过系统代理
  proxyHost: '',                         // frpc 代理 Host
  proxyPort: 0,                          // frpc 代理端口
};
let appSettings = {};                    // 当前生效的应用设置

const frpcProcesses = new Map();         // tunnelId -> { child, configPath }
const frpcLogs = new Map();              // tunnelId -> 最近日志行（排错用）

let pollTimer = null;                    // 轮询定时器
let pollDelay = 3000;                    // 当前轮询间隔（成功 3s，失败按指数退避）
let pollFailCount = 0;                   // 连续失败次数
let quitting = false;                    // 退出标记

// ===================== 基础工具 =====================

function log(...args) {
  console.log('[WeaveNet]', ...args);
}

// 持久化设置（frpc 路径、应用设置）到 userData 目录；服务器地址为内置常量，不落盘
function stateFile() {
  return path.join(app.getPath('userData'), 'state.json');
}

function loadState() {
  appSettings = { ...DEFAULT_SETTINGS };
  try {
    const raw = fs.readFileSync(stateFile(), 'utf8');
    const data = JSON.parse(raw);
    if (typeof data.frpcPath === 'string') customFrpcPath = data.frpcPath;
    if (data && typeof data.settings === 'object' && data.settings) {
      appSettings = { ...DEFAULT_SETTINGS, ...data.settings };
    }
  } catch (err) {
    // 首次运行无状态文件，使用默认值
  }
}

function saveState() {
  try {
    fs.mkdirSync(path.dirname(stateFile()), { recursive: true });
    fs.writeFileSync(
      stateFile(),
      JSON.stringify({ frpcPath: customFrpcPath, settings: appSettings }, null, 2),
      'utf8'
    );
  } catch (err) {
    log('保存状态失败', err.message);
  }
}

// 向渲染进程推送事件
function sendToRenderer(channel, payload) {
  if (mainWindow && !mainWindow.isDestroyed()) {
    mainWindow.webContents.send(channel, payload);
  }
}

// IPC 统一包装：handler 成功返回 { ok: true, data }，失败返回 { ok: false, error }
function handleIpc(handler) {
  return async (event, ...args) => {
    try {
      const data = await handler(event, ...args);
      return { ok: true, data };
    } catch (err) {
      log('IPC 调用失败:', err.message);
      return { ok: false, error: err.message || '未知错误' };
    }
  };
}

// ===================== 面板 API 网络请求 =====================

// 发起 JSON 请求，返回 { status, json }
function apiRequest(method, apiPath, body) {
  return new Promise((resolve, reject) => {
    let url;
    try {
      url = new URL(serverUrl);
    } catch (err) {
      reject(new Error('服务器地址无效，请在设置中检查'));
      return;
    }
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      reject(new Error('服务器地址仅支持 http 或 https 协议'));
      return;
    }
    const isHttps = url.protocol === 'https:';
    const mod = isHttps ? https : http;
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers.Authorization = 'Bearer ' + token;

    const req = mod.request(
      {
        hostname: url.hostname,
        port: url.port || (isHttps ? 443 : 80),
        path: apiPath,
        method,
        headers,
        timeout: 10000,
      },
      (res) => {
        let raw = '';
        res.setEncoding('utf8');
        res.on('data', (chunk) => {
          raw += chunk;
        });
        res.on('end', () => {
          let json = null;
          try {
            json = JSON.parse(raw);
          } catch (err) {
            json = null;
          }
          resolve({ status: res.statusCode, json });
        });
      }
    );
    req.on('error', (err) => reject(err));
    req.on('timeout', () => {
      req.destroy(new Error('请求超时，请检查网络与服务器地址'));
    });
    if (body !== undefined && body !== null) req.write(JSON.stringify(body));
    req.end();
  });
}

// 解析统一响应：成功 code === 0 返回 data；2xx 无 body（如 204）视为成功；其余抛出带 message 的错误
function unwrapResponse(res) {
  if (res.json && res.json.code === 0) {
    return res.json.data !== undefined ? res.json.data : {};
  }
  if (res.status >= 200 && res.status < 300 && !res.json) {
    return {};
  }
  const message = (res.json && res.json.message) || `请求失败（HTTP ${res.status || '未知'}）`;
  const err = new Error(message);
  err.httpCode = res.status || 0;
  err.businessCode = res.json ? res.json.business_code : null;
  throw err;
}

// ===================== frpc 子进程管理 =====================

// 解析 frpc 日志行：2026-08-15 10:00:00.000 [I] [xxx.go:200] 消息
function parseFrpcLogLine(raw) {
  const line = String(raw).trim();
  if (!line) return null;
  const m = /^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?)\s*\[([IWED])\]\s*(.*)$/.exec(line);
  if (m) {
    const map = { I: 'info', W: 'warn', E: 'error', D: 'debug' };
    return { time: m[1], level: map[m[2]] || 'info', msg: m[3] || line };
  }
  return { time: '', level: 'info', msg: line };
}

// 查找 frpc.exe：优先用户自定义路径，其次打包资源目录，最后开发目录
function findFrpcPath() {
  if (customFrpcPath) {
    try {
      if (fs.existsSync(customFrpcPath)) return customFrpcPath;
    } catch (err) {
      // 忽略
    }
  }
  const candidates = [
    path.join(process.resourcesPath || __dirname, 'bin', 'frpc.exe'),
    path.join(process.resourcesPath || __dirname, 'frpc.exe'),
    path.join(__dirname, 'resources', 'frpc.exe'),
    path.join(__dirname, 'bin', 'frpc.exe'),
  ];
  for (const candidate of candidates) {
    try {
      if (fs.existsSync(candidate)) return candidate;
    } catch (err) {
      // 忽略
    }
  }
  return '';
}

// Windows 下强制终止进程树（frpc 可能派生线程）
function killProcessTree(pid) {
  return new Promise((resolve) => {
    try {
      const killer = spawn('taskkill', ['/pid', String(pid), '/T', '/F'], { windowsHide: true });
      killer.on('close', () => resolve());
      killer.on('error', () => resolve());
    } catch (err) {
      resolve();
    }
  });
}

// 停止全部 frpc 子进程（退出登录 / 应用退出时调用）
async function stopAllFrpc() {
  const ids = Array.from(frpcProcesses.keys());
  await Promise.all(
    ids.map(async (tunnelId) => {
      const rec = frpcProcesses.get(tunnelId);
      if (rec) await killProcessTree(rec.child.pid);
    })
  );
  frpcProcesses.clear();
}

// ===================== 状态轮询（3s + 指数退避重连） =====================

function startStatusPolling() {
  stopStatusPolling();
  pollDelay = 3000;
  pollFailCount = 0;
  schedulePoll();
}

function stopStatusPolling() {
  if (pollTimer) {
    clearTimeout(pollTimer);
    pollTimer = null;
  }
}

function schedulePoll() {
  if (!token || quitting) return;
  pollTimer = setTimeout(runPoll, pollDelay);
}

async function runPoll() {
  if (!token) return;
  try {
    const res = await apiRequest('GET', '/api/tunnels');
    const data = unwrapResponse(res);
    // 兼容两种返回形态：data 为数组，或 data.tunnels 为数组
    const tunnels = Array.isArray(data)
      ? data
      : Array.isArray(data && data.tunnels)
        ? data.tunnels
        : [];
    pollFailCount = 0;
    pollDelay = 3000;
    sendToRenderer('tunnel-status', { connected: true, tunnels, updatedAt: Date.now() });
  } catch (err) {
    pollFailCount += 1;
    if (err.httpCode === 401) {
      // 登录会话过期：停止轮询并通知渲染进程回登录页
      stopStatusPolling();
      token = '';
      currentUser = null;
      sendToRenderer('session-expired', { message: err.message });
      return;
    }
    // 指数退避：1s 起，翻倍至最大 60s，自动重试
    pollDelay = Math.min(60000, 1000 * Math.pow(2, Math.min(pollFailCount - 1, 6)));
    sendToRenderer('tunnel-status', { connected: false, error: err.message });
  }
  schedulePoll();
}

// ===================== IPC 处理器 =====================

// 登录：POST /api/client/login
ipcMain.handle(
  'login',
  handleIpc(async (_event, payload) => {
    if (!payload || !payload.username || !payload.password) {
      throw new Error('请输入用户名和密码');
    }
    const res = await apiRequest('POST', '/api/client/login', {
      username: payload.username,
      password: payload.password,
    });
    const data = unwrapResponse(res);
    if (!data || !data.token) throw new Error('登录响应缺少令牌');
    token = data.token;
    currentUser = data.user || null;
    log('用户登录成功:', currentUser && currentUser.username);
    startStatusPolling();
    return data;
  })
);

// 退出登录：清令牌、停轮询、杀全部 frpc
ipcMain.handle(
  'logout',
  handleIpc(async () => {
    stopStatusPolling();
    await stopAllFrpc();
    token = '';
    currentUser = null;
    log('用户已退出登录');
    return { ok: true };
  })
);

// 拉取隧道列表：GET /api/client/tunnels
ipcMain.handle(
  'get-tunnels',
  handleIpc(async () => {
    const res = await apiRequest('GET', '/api/client/tunnels');
    return unwrapResponse(res);
  })
);

// 获取面板生成的 frpc.toml 内容（只读预览）：POST /api/client/config
ipcMain.handle(
  'get-tunnel-config',
  handleIpc(async (_event, tunnelId) => {
    const res = await apiRequest('POST', '/api/client/config', {
      tunnel_id: Number(tunnelId),
    });
    const data = unwrapResponse(res);
    return typeof data === 'string' ? data : (data && data.config) || '';
  })
);

// 启动隧道：面板生成配置 -> 写入临时目录 -> spawn frpc.exe
ipcMain.handle(
  'start-tunnel',
  handleIpc(async (_event, tunnelId) => {
    tunnelId = Number(tunnelId);
    if (frpcProcesses.has(tunnelId)) {
      return { alreadyRunning: true, pid: frpcProcesses.get(tunnelId).child.pid };
    }
    const frpc = findFrpcPath();
    if (!frpc) {
      throw new Error('未找到 frpc.exe，请将其放入 client/bin 或 client/resources 目录，或在设置中指定路径');
    }
    // 由面板签发完整配置（含 u_ 鉴权 Token，客户端无法本地计算）
    const res = await apiRequest('POST', '/api/client/config', { tunnel_id: tunnelId });
    const data = unwrapResponse(res);
    const configText = typeof data === 'string' ? data : (data && data.config) || '';
    if (!configText) throw new Error('面板返回的隧道配置为空');

    // 写入临时目录：%TEMP%/weavenet/frpc-<tunnelId>.toml
    const tmpDir = path.join(os.tmpdir(), 'weavenet');
    fs.mkdirSync(tmpDir, { recursive: true });
    const configPath = path.join(tmpDir, `frpc-${tunnelId}.toml`);
    fs.writeFileSync(configPath, configText, 'utf8');

    // 启动 frpc 子进程，工作目录设为配置所在目录
    const child = spawn(frpc, ['-c', configPath], {
      cwd: tmpDir,
      windowsHide: true,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    frpcProcesses.set(tunnelId, { child, configPath });

    // 收集子进程日志（结构化保留最近 300 行，退出时把尾部推给渲染进程）
    const entries = [];
    frpcLogs.set(tunnelId, entries);
    const pushLog = (chunk) => {
      const parts = String(chunk).split(/\r?\n/);
      for (const part of parts) {
        const parsed = parseFrpcLogLine(part);
        if (parsed) entries.push({ tunnelId, ...parsed });
      }
      if (entries.length > 300) entries.splice(0, entries.length - 300);
    };
    if (child.stdout) child.stdout.on('data', pushLog);
    if (child.stderr) child.stderr.on('data', pushLog);

    child.on('error', (err) => {
      log('frpc 启动失败:', err.message);
      frpcProcesses.delete(tunnelId);
      sendToRenderer('tunnel-exited', { tunnelId, code: null, log: err.message });
    });
    child.on('exit', (code, signal) => {
      log('frpc 进程退出:', tunnelId, 'code =', code, 'signal =', signal);
      frpcProcesses.delete(tunnelId);
      const tail = (frpcLogs.get(tunnelId) || []).slice(-5);
      frpcLogs.delete(tunnelId);
      sendToRenderer('tunnel-exited', {
        tunnelId,
        code,
        signal,
        log: tail.map((x) => `[${x.time}] [${x.level}] ${x.msg}`).join('\n'),
      });
    });

    log('隧道启动中:', tunnelId, 'pid =', child.pid, '配置 =', configPath);
    return { alreadyRunning: false, pid: child.pid, configPath };
  })
);

// 停止隧道：终止对应 frpc 子进程
ipcMain.handle(
  'stop-tunnel',
  handleIpc(async (_event, tunnelId) => {
    tunnelId = Number(tunnelId);
    const rec = frpcProcesses.get(tunnelId);
    if (!rec) return { alreadyStopped: true };
    await killProcessTree(rec.child.pid);
    frpcProcesses.delete(tunnelId);
    log('隧道已停止:', tunnelId);
    return { alreadyStopped: false };
  })
);

// 设置 frpc 可执行文件路径（空串表示自动查找）
ipcMain.handle(
  'set-frpc-path',
  handleIpc(async (_event, filePath) => {
    const value = String(filePath || '').trim();
    if (value) {
      try {
        if (!fs.existsSync(value)) throw new Error('指定的 frpc.exe 不存在');
      } catch (err) {
        if (err instanceof Error && err.message === '指定的 frpc.exe 不存在') throw err;
        throw new Error('无法访问指定路径');
      }
    }
    customFrpcPath = value;
    saveState();
    return { frpcPath: value };
  })
);

// 读取客户端配置信息（frpc 路径与是否存在、版本）
ipcMain.handle(
  'get-config',
  handleIpc(async () => ({
    frpcPath: findFrpcPath(),
    frpcFound: !!findFrpcPath(),
    version: app.getVersion(),
  }))
);

// 获取当前登录会话（仅返回用户信息，令牌绝不下发渲染进程）
ipcMain.handle(
  'get-session',
  handleIpc(async () => ({ loggedIn: !!currentUser, user: currentUser }))
);

// 创建隧道：POST /api/tunnels
ipcMain.handle(
  'create-tunnel',
  handleIpc(async (_event, payload) => {
    if (!payload || !payload.name) throw new Error('请填写隧道名称');
    const res = await apiRequest('POST', '/api/tunnels', payload);
    return unwrapResponse(res);
  })
);

// 删除隧道：DELETE /api/tunnels/{id}，同时停止本地 frpc 进程
ipcMain.handle(
  'delete-tunnel',
  handleIpc(async (_event, tunnelId) => {
    tunnelId = Number(tunnelId);
    const res = await apiRequest('DELETE', `/api/tunnels/${tunnelId}`);
    const data = unwrapResponse(res);
    const rec = frpcProcesses.get(tunnelId);
    if (rec) {
      await killProcessTree(rec.child.pid);
      frpcProcesses.delete(tunnelId);
    }
    frpcLogs.delete(tunnelId);
    return data;
  })
);

// 近 N 日流量统计：GET /api/stats/traffic?days=N
ipcMain.handle(
  'get-traffic',
  handleIpc(async (_event, days) => {
    const n = Math.max(1, Math.min(31, Number(days) || 7));
    const res = await apiRequest('GET', `/api/stats/traffic?days=${n}`);
    return unwrapResponse(res);
  })
);

// 套餐额度：GET /api/user/quota
ipcMain.handle(
  'get-quota',
  handleIpc(async () => unwrapResponse(await apiRequest('GET', '/api/user/quota')))
);

// 公告列表：GET /api/announcements（公开接口，未登录也可拉取）
ipcMain.handle(
  'get-announcements',
  handleIpc(async () => unwrapResponse(await apiRequest('GET', '/api/announcements')))
);

// 每日签到：POST /api/signin
ipcMain.handle(
  'signin',
  handleIpc(async () => unwrapResponse(await apiRequest('POST', '/api/signin')))
);

// 签到状态：GET /api/signin/status
ipcMain.handle(
  'get-signin-status',
  handleIpc(async () => unwrapResponse(await apiRequest('GET', '/api/signin/status')))
);

// 节点列表：GET /api/nodes
ipcMain.handle(
  'get-nodes',
  handleIpc(async () => unwrapResponse(await apiRequest('GET', '/api/nodes')))
);

// 本地 frpc 运行日志：合并全部隧道日志，支持按隧道过滤
ipcMain.handle(
  'get-tunnel-logs',
  handleIpc(async (_event, tunnelId, maxLines) => {
    const limit = Math.max(1, Number(maxLines) || 200);
    const out = [];
    for (const [id, entries] of frpcLogs) {
      if (tunnelId != null && Number(id) !== Number(tunnelId)) continue;
      out.push(...entries);
    }
    return out.slice(-limit);
  })
);

// 读取应用设置（本地持久化）
ipcMain.handle(
  'get-app-settings',
  handleIpc(async () => ({ ...appSettings }))
);

// 保存应用设置（本地持久化）
ipcMain.handle(
  'save-app-settings',
  handleIpc(async (_event, settings) => {
    if (!settings || typeof settings !== 'object') throw new Error('设置格式无效');
    appSettings = { ...appSettings, ...settings };
    saveState();
    return { ...appSettings };
  })
);

// 打开 frpc 配置临时目录
ipcMain.handle(
  'open-log-dir',
  handleIpc(async () => {
    const dir = path.join(os.tmpdir(), 'weavenet');
    fs.mkdirSync(dir, { recursive: true });
    const err = await shell.openPath(dir);
    if (err) throw new Error(err);
    return { dir };
  })
);

// 选择 frpc.exe 文件（文件选择对话框）
ipcMain.handle(
  'pick-frpc-path',
  handleIpc(async () => {
    const result = await dialog.showOpenDialog(mainWindow, {
      title: '选择 frpc 可执行文件',
      filters: [{ name: '可执行文件', extensions: ['exe'] }],
      properties: ['openFile'],
    });
    if (result.canceled || !result.filePaths.length) return { canceled: true };
    return { canceled: false, path: result.filePaths[0] };
  })
);

// ===================== 窗口与生命周期 =====================

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 780,
    minWidth: 980,
    minHeight: 640,
    title: 'WeaveNet 织网穿透',
    autoHideMenuBar: true,
    backgroundColor: '#f0f9ff',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });
  mainWindow.loadFile(path.join(__dirname, 'dist', 'index.html'));
  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

app.whenReady().then(() => {
  loadState();
  createWindow();
  log('客户端已启动，面板地址:', serverUrl);

  app.on('activate', () => {
    // macOS 上点击 Dock 图标时重新创建窗口
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});

app.on('before-quit', () => {
  quitting = true;
  stopStatusPolling();
  // 应用退出时尽量终止 frpc 子进程
  for (const rec of frpcProcesses.values()) {
    try {
      rec.child.kill();
    } catch (err) {
      // 忽略
    }
  }
});
