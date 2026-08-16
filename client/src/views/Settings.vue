<template>
  <div class="settings">
    <div class="settings-grid">
      <!-- 通用设置 -->
      <div class="card">
        <div class="card-head">
          <h3>通用设置</h3>
        </div>
        <div class="setting-list">
          <label class="setting-row">
            <span class="setting-info">
              <strong>启动时自动检测更新</strong>
              <small>应用启动时检查是否有新版本</small>
            </span>
            <input type="checkbox" v-model="settings.autoCheckUpdate" @change="save" />
          </label>
          <label class="setting-row">
            <span class="setting-info">
              <strong>关闭窗口最小化到托盘</strong>
              <small>点击关闭按钮时隐藏到系统托盘而非退出</small>
            </span>
            <input type="checkbox" v-model="settings.minimizeToTray" @change="save" />
          </label>
          <label class="setting-row">
            <span class="setting-info">
              <strong>进程守护</strong>
              <small>监控隧道进程，意外退出时自动重启</small>
            </span>
            <input type="checkbox" v-model="settings.processGuard" @change="save" />
          </label>
          <label class="setting-row">
            <span class="setting-info">
              <strong>修改隧道后自动重启映射</strong>
              <small>隧道配置变化后自动重启对应进程</small>
            </span>
            <input type="checkbox" v-model="settings.autoRestartOnChange" @change="save" />
          </label>
          <div class="setting-row">
            <span class="setting-info">
              <strong>frpc 日志输出等级</strong>
              <small>调整底层 frpc 内核的日志详细程度</small>
            </span>
            <select v-model="settings.frpcLogLevel" @change="save" class="setting-select">
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>
          </div>
        </div>
      </div>

      <!-- 个性化 -->
      <div class="card">
        <div class="card-head">
          <h3>个性化</h3>
        </div>
        <div class="setting-list">
          <div class="setting-row">
            <span class="setting-info">
              <strong>主题</strong>
              <small>跟随系统或手动切换明暗</small>
            </span>
            <select v-model="settings.theme" @change="save" class="setting-select">
              <option value="system">跟随系统</option>
              <option value="light">浅色</option>
              <option value="dark">深色</option>
            </select>
          </div>
          <div class="setting-row">
            <span class="setting-info">
              <strong>侧边栏显示样式</strong>
              <small>选择导航栏的呈现方式</small>
            </span>
            <select v-model="settings.sidebarStyle" @change="save" class="setting-select">
              <option value="default">默认</option>
              <option value="compact">紧凑</option>
            </select>
          </div>
          <label class="setting-row">
            <span class="setting-info">
              <strong>显示顶部标题栏</strong>
              <small>控制窗口顶部标题栏显示开关</small>
            </span>
            <input type="checkbox" v-model="settings.showTitleBar" @change="save" />
          </label>
          <label class="setting-row">
            <span class="setting-info">
              <strong>操作提示音效</strong>
              <small>操作成功或失败时播放提示音</small>
            </span>
            <input type="checkbox" v-model="settings.soundEnabled" @change="save" />
          </label>
          <div class="setting-row">
            <span class="setting-info">
              <strong>背景</strong>
              <small>自定义窗口背景图片、视频或文件夹轮播</small>
            </span>
            <select v-model="settings.backgroundType" @change="save" class="setting-select">
              <option value="none">无</option>
              <option value="image">图片</option>
              <option value="video">视频</option>
              <option value="folder">文件夹轮播</option>
            </select>
          </div>
          <div class="setting-row" v-if="settings.backgroundType !== 'none'">
            <input v-model="settings.backgroundValue" type="text" class="setting-input" spellcheck="false" placeholder="图片/视频路径或文件夹路径" @change="save" />
          </div>
        </div>
      </div>

      <!-- 网络设置 -->
      <div class="card">
        <div class="card-head">
          <h3>网络设置</h3>
        </div>
        <div class="setting-list">
          <label class="setting-row">
            <span class="setting-info">
              <strong>绕过系统代理</strong>
              <small>隧道流量不经过系统代理设置</small>
            </span>
            <input type="checkbox" v-model="settings.bypassProxy" @change="save" />
          </label>
          <div class="setting-row">
            <span class="setting-info">
              <strong>frpc 代理 Host</strong>
              <small>frpc 底层代理服务器地址（留空不启用）</small>
            </span>
            <input v-model="settings.proxyHost" type="text" class="setting-input" spellcheck="false" placeholder="如 127.0.0.1" @change="save" />
          </div>
          <div class="setting-row">
            <span class="setting-info">
              <strong>frpc 代理端口</strong>
              <small>代理服务器端口</small>
            </span>
            <input v-model.number="settings.proxyPort" type="number" class="setting-input" min="1" max="65535" placeholder="如 7890" @change="save" />
          </div>
        </div>
      </div>

      <!-- frpc 与版本 -->
      <div class="card">
        <div class="card-head">
          <h3>frpc 与版本</h3>
        </div>
        <div class="setting-list">
          <div class="setting-row">
            <span class="setting-info">
              <strong>frpc 可执行文件路径</strong>
              <small>{{ frpcHint }}</small>
            </span>
            <div class="row-inline">
              <input v-model="frpcPath" type="text" class="setting-input" spellcheck="false" placeholder="留空自动查找 client/bin 或 client/resources" />
              <button class="btn btn-ghost btn-sm" @click="pickFrpc">浏览</button>
            </div>
          </div>
          <div class="setting-row">
            <span class="setting-info">
              <strong>客户端版本</strong>
              <small>当前安装的 WeaveNet 客户端版本</small>
            </span>
            <span class="version-badge">v{{ store.version || '0.2.0' }}</span>
          </div>
        </div>
      </div>

      <!-- 关于 -->
      <div class="card about-card">
        <div class="card-head">
          <h3>关于</h3>
        </div>
        <div class="about-body">
          <div class="about-brand">
            <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg" class="about-logo">
              <defs>
                <linearGradient id="about-grad" x1="8" y1="8" x2="56" y2="56" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#0ea5e9" />
                  <stop offset="1" stop-color="#06b6d4" />
                </linearGradient>
              </defs>
              <g stroke="url(#about-grad)" stroke-width="3" stroke-linecap="round">
                <line x1="16" y1="16" x2="32" y2="32" />
                <line x1="32" y1="32" x2="48" y2="16" />
                <line x1="16" y1="16" x2="16" y2="48" />
                <line x1="48" y1="16" x2="48" y2="48" />
                <line x1="16" y1="48" x2="32" y2="32" />
                <line x1="32" y1="32" x2="48" y2="48" />
              </g>
              <g fill="url(#about-grad)">
                <circle cx="16" cy="16" r="5" />
                <circle cx="48" cy="16" r="5" />
                <circle cx="16" cy="48" r="5" />
                <circle cx="48" cy="48" r="5" />
                <circle cx="32" cy="32" r="6" />
              </g>
            </svg>
            <div>
              <strong>WeaveNet 织网穿透</strong>
              <small>v{{ store.version || '0.2.0' }} · CQFISH&喵酱出品</small>
            </div>
          </div>
          <p class="about-desc">免配置桌面客户端，登录即可自动同步隧道配置，开箱即用。底层基于 frpc 内核完成隧道内网穿透。</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import { api } from '../api';
import { store } from '../store';

const settings = ref({ ...store.appSettings });
const frpcPath = ref('');
const frpcHint = ref('');

async function save() {
  store.appSettings = Object.assign({}, store.appSettings, settings.value);
  try {
    await api.saveAppSettings(settings.value);
  } catch (err) {
    // 忽略
  }
}

async function pickFrpc() {
  // 由主进程提供文件选择对话框
  try {
    const res = await window.weavenet.invoke('pick-frpc-path');
    if (res && res.ok && res.data) {
      frpcPath.value = res.data;
      await saveFrpc();
    }
  } catch (err) {
    alert(err.message);
  }
}

async function saveFrpc() {
  try {
    await api.setFrpcPath(frpcPath.value.trim());
    const cfg = await api.getConfig();
    frpcHint.value = cfg.frpcFound ? `已找到 frpc：${cfg.frpcPath}` : '未找到 frpc.exe，请指定路径';
    store.frpcFound = !!cfg.frpcFound;
  } catch (err) {
    alert(err.message);
  }
}

onMounted(async () => {
  try {
    const st = await api.getAppSettings();
    if (st) {
      settings.value = Object.assign({}, store.appSettings, st);
      store.appSettings = { ...settings.value };
    }
  } catch (err) {
    // 忽略
  }
  try {
    const cfg = await api.getConfig();
    frpcPath.value = cfg.frpcPath || '';
    store.version = cfg.version || store.version;
    frpcHint.value = cfg.frpcFound ? `已找到 frpc：${cfg.frpcPath}` : '未找到 frpc.exe，请指定路径';
  } catch (err) {
    // 忽略
  }
});

watch(() => store.appSettings, (val) => {
  settings.value = { ...val };
}, { deep: true });
</script>

<style scoped>
.settings {
  max-width: 1000px;
  margin: 0 auto;
}

.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}

.card-head h3 {
  font-size: 15px;
}

.setting-list {
  display: flex;
  flex-direction: column;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 11px 2px;
  border-bottom: 1px dashed var(--border);
}

.setting-row:last-child {
  border-bottom: none;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.setting-info strong {
  font-size: 13.5px;
}

.setting-info small {
  font-size: 12px;
  color: var(--text-faint);
}

.setting-row input[type='checkbox'] {
  width: 17px;
  height: 17px;
  accent-color: var(--primary);
  flex-shrink: 0;
}

.setting-select,
.setting-input {
  padding: 7px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-strong);
  background: var(--surface-strong);
  color: var(--text);
  font-size: 12.5px;
  font-family: inherit;
  outline: none;
  flex-shrink: 0;
}

.setting-select:focus,
.setting-input:focus {
  border-color: var(--primary);
}

.setting-input {
  width: 190px;
}

.row-inline {
  display: flex;
  gap: 8px;
  align-items: center;
}

.version-badge {
  padding: 4px 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.16), rgba(6, 182, 212, 0.14));
  border: 1px solid rgba(14, 165, 233, 0.4);
  color: var(--primary-deep);
  font-size: 12.5px;
  font-weight: 700;
}

.theme-dark .version-badge {
  color: #7dd3fc;
}

.about-card {
  grid-column: span 2;
}

.about-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.about-brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.about-logo {
  width: 44px;
  height: 44px;
}

.about-brand div {
  display: flex;
  flex-direction: column;
  line-height: 1.35;
}

.about-brand strong {
  font-size: 15.5px;
}

.about-brand small {
  font-size: 12.5px;
  color: var(--text-faint);
}

.about-desc {
  font-size: 13px;
  color: var(--text-sub);
  line-height: 1.7;
}

@media (max-width: 900px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }

  .about-card {
    grid-column: span 1;
  }

  .setting-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .setting-input {
    width: 100%;
  }
}
</style>
