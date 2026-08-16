'use strict';

const { contextBridge, ipcRenderer } = require('electron');

// 渲染进程统一通过 weavenet.invoke(channel, ...args) 调用主进程
// 主进程每个 handler 统一返回 { ok: true, data } 或 { ok: false, error }
contextBridge.exposeInMainWorld('weavenet', {
  // 泛化调用：src/api.js 全部 API 的底层入口
  invoke: (channel, ...args) => ipcRenderer.invoke(channel, ...args),

  // 订阅实时状态推送（每 3 秒，含隧道在线/流量信息）
  onStatusUpdate: (callback) => {
    ipcRenderer.removeAllListeners('tunnel-status');
    ipcRenderer.on('tunnel-status', (_event, data) => callback(data));
  },

  // 订阅会话过期事件（登录失效时主进程推送）
  onSessionExpired: (callback) => {
    ipcRenderer.removeAllListeners('session-expired');
    ipcRenderer.on('session-expired', (_event, data) => callback(data));
  },

  // 订阅隧道进程退出事件
  onTunnelExited: (callback) => {
    ipcRenderer.removeAllListeners('tunnel-exited');
    ipcRenderer.on('tunnel-exited', (_event, data) => callback(data));
  },
});
