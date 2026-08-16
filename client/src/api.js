// 渲染进程与主进程通信的统一封装
// 所有 IPC 调用返回 { ok, data | error }，这里负责解包并抛出干净错误

async function call(channel, ...args) {
  const res = await window.weavenet.invoke(channel, ...args);
  if (res && res.ok === false) {
    throw new Error(res.error || '操作失败');
  }
  return res && res.data;
}

export const api = {
  // 会话
  login: (payload) => call('login', payload),
  logout: () => call('logout'),
  getSession: () => call('get-session'),

  // 配置
  getConfig: () => call('get-config'),
  setServerUrl: (url) => call('set-server-url', url),
  setFrpcPath: (p) => call('set-frpc-path', p),

  // 隧道
  getTunnels: () => call('get-tunnels'),
  createTunnel: (data) => call('create-tunnel', data),
  deleteTunnel: (id) => call('delete-tunnel', id),
  startTunnel: (id) => call('start-tunnel', id),
  stopTunnel: (id) => call('stop-tunnel', id),
  getTunnelConfig: (id) => call('get-tunnel-config', id),

  // 数据
  getTraffic: (days = 7) => call('get-traffic', days),
  getQuota: () => call('get-quota'),
  getAnnouncements: () => call('get-announcements'),
  signin: () => call('signin'),
  getSigninStatus: () => call('get-signin-status'),
  getNodes: () => call('get-nodes'),

  // 日志（本地 frpc 运行日志）
  getTunnelLogs: (tunnelId, maxLines) => call('get-tunnel-logs', tunnelId, maxLines),

  // 应用设置（本地持久化）
  getAppSettings: () => call('get-app-settings'),
  saveAppSettings: (settings) => call('save-app-settings', settings),
  openLogDir: () => call('open-log-dir'),

  // 事件订阅
  onStatusUpdate: (cb) => window.weavenet.onStatusUpdate(cb),
  onSessionExpired: (cb) => window.weavenet.onSessionExpired(cb),
  onTunnelExited: (cb) => window.weavenet.onTunnelExited(cb),
};
