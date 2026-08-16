<template>
  <div class="layout">
    <header class="topbar">
      <div class="topbar-left">
        <router-link to="/admin/dashboard" class="brand">
          <svg width="32" height="32" viewBox="0 0 64 64">
            <defs>
              <linearGradient id="lg-adminbar" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0" stop-color="#0ea5e9" />
                <stop offset="1" stop-color="#06b6d4" />
              </linearGradient>
            </defs>
            <rect width="64" height="64" rx="16" fill="url(#lg-adminbar)" />
            <g fill="none" stroke="#ffffff" stroke-width="4" stroke-linecap="round">
              <circle cx="24" cy="24" r="7" />
              <circle cx="42" cy="42" r="7" />
              <path d="M30 30c4 4 5 7 5 12" />
              <path d="M24 31c-1 6-2 8-6 11" />
            </g>
          </svg>
          <span class="brand-text">WeaveNet 管理后台</span>
          <span class="admin-badge">ADMIN</span>
        </router-link>
      </div>
      <div class="topbar-right">
        <n-dropdown trigger="click" :options="userMenuOptions" @select="handleUserMenu">
          <div class="user-trigger">
            <div class="user-avatar">{{ avatarText }}</div>
            <span class="user-name">{{ userStore.displayName }}</span>
            <svg-icon name="chevronDown" :size="14" class="chevron" />
          </div>
        </n-dropdown>
      </div>
    </header>
    <div class="layout-body">
      <aside class="glass-card sidebar">
        <nav class="menu-nav">
          <router-link
            v-for="item in menus"
            :key="item.path"
            :to="item.path"
            class="menu-item"
            :class="{ active: route.path === item.path }"
          >
            <span class="mi-icon">
              <svg-icon :name="item.icon" :size="18" />
            </span>
            <span>{{ item.label }}</span>
          </router-link>
        </nav>
        <div class="sidebar-foot">
          <router-link to="/panel/dashboard" class="user-entry">
            <svg-icon name="home" :size="15" />
            <span>返回用户面板</span>
          </router-link>
        </div>
      </aside>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, useDialog, NDropdown } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { useUserStore } from '@/store/user'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const userStore = useUserStore()

const menus = [
  { path: '/admin/dashboard', label: '数据看板', icon: 'dashboard' },
  { path: '/admin/users', label: '用户管理', icon: 'users' },
  { path: '/admin/nodes', label: '节点管理', icon: 'node' },
  { path: '/admin/tunnels', label: '隧道管理', icon: 'tunnel' },
  { path: '/admin/plans', label: '套餐配置', icon: 'plan' },
  { path: '/admin/announcements', label: '公告管理', icon: 'announcement' },
  { path: '/admin/config', label: '系统配置', icon: 'config' },
  { path: '/admin/logs', label: '操作日志', icon: 'logs' }
]

const avatarText = computed(() => (userStore.displayName || 'A').slice(0, 1).toUpperCase())

const userMenuOptions = [
  { label: '返回用户面板', key: 'panel' },
  { label: '退出登录', key: 'logout' }
]

function handleUserMenu(key) {
  if (key === 'panel') {
    router.push('/panel/dashboard')
  } else if (key === 'logout') {
    dialog.warning({
      title: '退出登录',
      content: '确定要退出管理后台吗？',
      positiveText: '退出',
      negativeText: '取消',
      onPositiveClick: async () => {
        await userStore.logout()
        message.success('已安全退出')
        router.push('/admin/login')
      }
    })
  }
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  height: 60px;
  padding: 0 22px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.55);
  -webkit-backdrop-filter: blur(16px);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.65);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}

.brand-text {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.admin-badge {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  letter-spacing: 1px;
  color: #b45309;
  background: rgba(245, 158, 11, 0.16);
  border: 1px solid rgba(245, 158, 11, 0.35);
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(255, 255, 255, 0.7);
}

.user-trigger:hover {
  background: rgba(14, 165, 233, 0.1);
}

.user-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #f59e0b, #f97316);
}

.user-name {
  font-size: 13px;
  color: #334155;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chevron {
  color: #64748b;
}

.layout-body {
  flex: 1;
  display: flex;
  gap: 16px;
  padding: 16px;
  width: 100%;
  max-width: 1440px;
  margin: 0 auto;
  align-items: flex-start;
}

.sidebar {
  width: 218px;
  flex-shrink: 0;
  padding: 12px;
  position: sticky;
  top: 76px;
}

.sidebar-foot {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.6);
}

.user-entry {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 12px;
  font-size: 13px;
  color: #64748b;
  text-decoration: none;
}

.user-entry:hover {
  background: rgba(14, 165, 233, 0.08);
  color: #0ea5e9;
}

.content {
  flex: 1;
  min-width: 0;
}

@media (max-width: 900px) {
  .sidebar {
    display: none;
  }
}
</style>
