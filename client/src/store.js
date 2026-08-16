import { reactive } from 'vue';

// 全局状态：登录态、用户信息、隧道、连接状态
export const store = reactive({
  user: null,
  loggedIn: false,
  serverUrl: 'http://127.0.0.1:8000',
  frpcFound: false,
  frpcPath: '',
  version: '',
  connected: true,
  tunnels: [],
  statusMap: new Map(),
  appSettings: {
    autoCheckUpdate: true,
    minimizeToTray: true,
    processGuard: true,
    frpcLogLevel: 'info',
    autoRestartOnChange: true,
    theme: 'system',
    sidebarStyle: 'default',
    showTitleBar: true,
    soundEnabled: true,
    backgroundType: 'none',
    backgroundValue: '',
    bypassProxy: false,
    proxyHost: '',
    proxyPort: 0,
  },
});

export function setUser(user) {
  store.user = user || null;
  store.loggedIn = !!user;
}

export function runtimeOf(tunnelId) {
  return store.statusMap.get(tunnelId) || {};
}

export function statusInfo(t) {
  const r = runtimeOf(t && t.id);
  if (typeof r.online === 'boolean') {
    if (r.online) return { text: '运行中', cls: 'status-running' };
    if (t && t.status === 'running') return { text: '离线', cls: 'status-offline' };
    return { text: '已停止', cls: 'status-stopped' };
  }
  return t && t.status === 'running'
    ? { text: '运行中', cls: 'status-running' }
    : { text: '已停止', cls: 'status-stopped' };
}

export function formatBytes(bytes) {
  const value = Number(bytes) || 0;
  if (value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let v = value;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(v >= 100 ? 0 : 1)} ${units[i]}`;
}
