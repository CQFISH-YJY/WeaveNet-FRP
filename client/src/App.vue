<template>
  <div class="app-shell" :class="{ 'theme-dark': isDark }">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-brand">
        <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg" class="brand-logo">
          <defs>
            <linearGradient id="side-grad" x1="8" y1="8" x2="56" y2="56" gradientUnits="userSpaceOnUse">
              <stop stop-color="#0ea5e9" />
              <stop offset="1" stop-color="#06b6d4" />
            </linearGradient>
          </defs>
          <g stroke="url(#side-grad)" stroke-width="3" stroke-linecap="round">
            <line x1="16" y1="16" x2="32" y2="32" />
            <line x1="32" y1="32" x2="48" y2="16" />
            <line x1="16" y1="16" x2="16" y2="48" />
            <line x1="48" y1="16" x2="48" y2="48" />
            <line x1="16" y1="48" x2="32" y2="32" />
            <line x1="32" y1="32" x2="48" y2="48" />
          </g>
          <g fill="url(#side-grad)">
            <circle cx="16" cy="16" r="5" />
            <circle cx="48" cy="16" r="5" />
            <circle cx="16" cy="48" r="5" />
            <circle cx="48" cy="48" r="5" />
            <circle cx="32" cy="32" r="6" />
          </g>
        </svg>
        <div class="brand-text">
          <strong>WeaveNet</strong>
          <span>织网穿透</span>
        </div>
      </div>

      <nav class="side-nav">
        <router-link to="/" class="nav-item" :class="{ active: route.path === '/' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 10.5 12 3l9 7.5" />
            <path d="M5 9.5V21h14V9.5" />
            <path d="M10 21v-6h4v6" />
          </svg>
          <span>首页</span>
        </router-link>
        <router-link to="/tunnels" class="nav-item" :class="{ active: route.path === '/tunnels' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 7h16" />
            <path d="M4 12h10" />
            <path d="M4 17h16" />
          </svg>
          <span>隧道</span>
        </router-link>
        <router-link to="/logs" class="nav-item" :class="{ active: route.path === '/logs' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 4h16v16H4z" />
            <path d="M8 9h8" />
            <path d="M8 13h8" />
            <path d="M8 17h5" />
          </svg>
          <span>日志</span>
        </router-link>
        <router-link to="/settings" class="nav-item" :class="{ active: route.path === '/settings' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1.03 1.56V21a2 2 0 1 1-4 0v-.09A1.7 1.7 0 0 0 9 19.36a1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.7 1.7 0 0 0 .34-1.87 1.7 1.7 0 0 0-1.56-1.03H3a2 2 0 1 1 0-4h.09A1.7 1.7 0 0 0 4.64 9a1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.7 1.7 0 0 0 1.87.34h.09a1.7 1.7 0 0 0 1.03-1.56V3a2 2 0 1 1 4 0v.09A1.7 1.7 0 0 0 15 4.64a1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.7 1.7 0 0 0-.34 1.87v.09a1.7 1.7 0 0 0 1.56 1.03H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.51.98Z" />
          </svg>
          <span>设置</span>
        </router-link>
      </nav>

      <div class="sidebar-foot">
        <div class="conn-pill" :class="{ offline: !store.connected }">
          <span class="conn-dot"></span>
          <span>{{ store.connected ? '已连接' : '重连中' }}</span>
        </div>
        <button class="user-card" @click="openLogin" v-if="!store.loggedIn">
          <span class="avatar">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="8" r="4" />
              <path d="M4 21c1.5-3.5 4.5-5 8-5s6.5 1.5 8 5" />
            </svg>
          </span>
          <span class="user-meta">
            <strong>未登录</strong>
            <small>点击登录使用完整功能</small>
          </span>
        </button>
        <div class="user-card" v-else>
          <span class="avatar logged">
            {{ (store.user && store.user.username || '?').charAt(0).toUpperCase() }}
          </span>
          <span class="user-meta">
            <strong>{{ store.user && store.user.username }}</strong>
            <small>{{ store.user && store.user.plan_name || '' }}</small>
          </span>
          <button class="logout-btn" title="退出登录" @click="logout">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <path d="M16 17l5-5-5-5" />
              <path d="M21 12H9" />
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- 主区域 -->
    <main class="main-area">
      <header class="topbar">
        <h1 class="page-title">{{ route.meta.title }}</h1>
        <div class="topbar-right">
          <div v-if="store.loggedIn" class="plan-chip">{{ store.user && store.user.plan_name }}</div>
          <button class="icon-btn" title="设置" @click="$router.push('/settings')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1.03 1.56V21a2 2 0 1 1-4 0v-.09A1.7 1.7 0 0 0 9 19.36a1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.7 1.7 0 0 0 .34-1.87 1.7 1.7 0 0 0-1.56-1.03H3a2 2 0 1 1 0-4h.09A1.7 1.7 0 0 0 4.64 9a1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.7 1.7 0 0 0 1.87.34h.09a1.7 1.7 0 0 0 1.03-1.56V3a2 2 0 1 1 4 0v.09A1.7 1.7 0 0 0 15 4.64a1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.7 1.7 0 0 0-.34 1.87v.09a1.7 1.7 0 0 0 1.56 1.03H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.51.98Z" />
            </svg>
          </button>
        </div>
      </header>

      <div v-if="!store.connected" class="conn-banner">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <path d="M12 9v4m0 4h.01" />
          <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z" />
        </svg>
        <span>连接中断，正在重试</span>
      </div>

      <section class="page-content">
        <router-view />
      </section>
    </main>

    <!-- 登录弹窗 -->
    <div v-if="showLoginModal" class="modal-mask" @click.self="showLoginModal = false">
      <div class="modal-panel glass">
        <div class="modal-head">
          <h2>登录账号</h2>
          <button class="icon-btn" @click="showLoginModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </div>
        <div class="field">
          <label>用户名 / 邮箱</label>
          <input v-model="loginForm.username" type="text" autocomplete="username" spellcheck="false" @keydown.enter="doLogin" />
        </div>
        <div class="field">
          <label>密码</label>
          <input v-model="loginForm.password" type="password" autocomplete="current-password" @keydown.enter="doLogin" />
        </div>
        <p v-if="loginError" class="form-error">{{ loginError }}</p>
        <button class="btn btn-primary btn-block" :disabled="loggingIn" @click="doLogin">
          {{ loggingIn ? '登录中...' : '登录' }}
        </button>
      </div>
    </div>

    <!-- 轻提示 -->
    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="{ error: toast.type === 'error' }">{{ toast.text }}</div>
    </transition>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { api } from './api';
import { store, setUser } from './store';

const route = useRoute();
const showLoginModal = ref(false);
const loggingIn = ref(false);
const loginError = ref('');
const loginForm = reactive({ username: '', password: '' });

const toast = reactive({ show: false, text: '', type: 'info' });
let toastTimer = null;

const isDark = computed(() => {
  if (store.appSettings.theme === 'dark') return true;
  if (store.appSettings.theme === 'light') return false;
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
});

function showToast(text, type = 'info') {
  toast.text = text;
  toast.type = type;
  toast.show = true;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { toast.show = false; }, 3200);
}

function openLogin() {
  loginError.value = '';
  showLoginModal.value = true;
}

async function doLogin() {
  const username = loginForm.username.trim();
  const password = loginForm.password;
  if (!username || !password) {
    loginError.value = '请填写完整的登录信息';
    return;
  }
  loggingIn.value = true;
  loginError.value = '';
  try {
    const data = await api.login({ username, password });
    setUser(data.user || null);
    showLoginModal.value = false;
    showToast('登录成功，欢迎回来');
    await refreshAll();
  } catch (err) {
    loginError.value = err.message || '登录失败';
  } finally {
    loggingIn.value = false;
  }
}

async function logout() {
  try {
    await api.logout();
  } catch (err) {
    // 忽略
  }
  setUser(null);
  store.tunnels = [];
  store.statusMap.clear();
  showToast('已退出登录');
}

async function refreshAll() {
  try {
    const data = await api.getTunnels();
    store.tunnels = (data && data.tunnels) || [];
    if (data && data.user) {
      setUser(Object.assign({}, store.user, data.user));
    }
  } catch (err) {
    // 忽略
  }
}

let statusTimer = null;

async function init() {
  try {
    const cfg = await api.getConfig();
    store.frpcFound = !!cfg.frpcFound;
    store.frpcPath = cfg.frpcPath || '';
    store.version = cfg.version || '';
  } catch (err) {
    // 忽略
  }
  try {
    const settings = await api.getAppSettings();
    if (settings) store.appSettings = Object.assign({}, store.appSettings, settings);
  } catch (err) {
    // 忽略
  }
  try {
    const session = await api.getSession();
    if (session && session.user) setUser(session.user);
  } catch (err) {
    // 忽略
  }

  api.onStatusUpdate((data) => {
    store.connected = !!data.connected;
    if (Array.isArray(data.tunnels) && data.tunnels.length > 0) {
      for (const item of data.tunnels) {
        store.statusMap.set(item.id, {
          online: !!item.online,
          connections: Number(item.connections) || 0,
          today_in: Number(item.today_in) || 0,
          today_out: Number(item.today_out) || 0,
        });
      }
    }
  });

  api.onSessionExpired((data) => {
    showToast((data && data.message) || '登录已过期，请重新登录', 'error');
    setUser(null);
  });

  api.onTunnelExited((data) => {
    if (data && data.tunnelId != null) {
      const existing = store.statusMap.get(data.tunnelId) || {};
      existing.online = false;
      store.statusMap.set(data.tunnelId, existing);
      const t = store.tunnels.find((x) => x.id === data.tunnelId);
      if (t) t.status = 'stopped';
      const codeText = data.code == null ? '启动失败' : `退出码 ${data.code}`;
      showToast(`隧道进程已退出（${codeText}）`, 'error');
    }
  });

  if (store.loggedIn) {
    statusTimer = setInterval(refreshAll, 15000);
  }
}

onMounted(init);
onUnmounted(() => {
  if (statusTimer) clearInterval(statusTimer);
});
</script>
