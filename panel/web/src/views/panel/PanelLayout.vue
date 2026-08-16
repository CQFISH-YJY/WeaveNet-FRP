<template>
  <div class="layout">
    <header class="topbar">
      <div class="topbar-left">
        <router-link to="/panel/dashboard" class="brand">
          <svg width="32" height="32" viewBox="0 0 64 64">
            <defs>
              <linearGradient id="lg-topbar" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0" stop-color="#0ea5e9" />
                <stop offset="1" stop-color="#06b6d4" />
              </linearGradient>
            </defs>
            <rect width="64" height="64" rx="16" fill="url(#lg-topbar)" />
            <g fill="none" stroke="#ffffff" stroke-width="4" stroke-linecap="round">
              <circle cx="24" cy="24" r="7" />
              <circle cx="42" cy="42" r="7" />
              <path d="M30 30c4 4 5 7 5 12" />
              <path d="M24 31c-1 6-2 8-6 11" />
            </g>
          </svg>
          <span class="brand-text">WeaveNet 织网穿透</span>
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
          <router-link to="/admin" class="admin-entry">
            <svg-icon name="shield" :size="15" />
            <span>管理后台入口</span>
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
import { useMessage, NDropdown } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { useUserStore } from '@/store/user'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const menus = [
  { path: '/panel/dashboard', label: '概览', icon: 'dashboard' },
  { path: '/panel/tunnels', label: '隧道管理', icon: 'tunnel' },
  { path: '/panel/stats', label: '流量统计', icon: 'stats' },
  { path: '/panel/points', label: '积分中心', icon: 'points' },
  { path: '/panel/announcements', label: '公告', icon: 'announcement' },
  { path: '/panel/tickets', label: '工单', icon: 'ticket' },
  { path: '/panel/profile', label: '个人中心', icon: 'profile' }
]

const avatarText = computed(() => (userStore.displayName || 'U').slice(0, 1).toUpperCase())

const userMenuOptions = computed(() => {
  const options = []
  if (userStore.isAdmin) {
    options.push({ label: '进入管理后台', key: 'admin' })
  }
  options.push(
    { label: '个人中心', key: 'profile' },
    { label: '退出登录', key: 'logout' }
  )
  return options
})

async function handleUserMenu(key) {
  if (key === 'profile') {
    router.push('/panel/profile')
  } else if (key === 'admin') {
    router.push('/admin/dashboard')
  } else if (key === 'logout') {
    await userStore.logout()
    message.success('已安全退出')
    router.push('/login')
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
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
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

.admin-entry {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 12px;
  font-size: 13px;
  color: #64748b;
  text-decoration: none;
}

.admin-entry:hover {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
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
