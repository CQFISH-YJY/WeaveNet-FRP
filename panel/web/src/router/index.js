import { createRouter, createWebHashHistory } from 'vue-router'
import { useUserStore } from '@/store/user'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { public: true, title: '登录' }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/Register.vue'),
    meta: { public: true, title: '注册' }
  },
  {
    path: '/forgot',
    name: 'forgot',
    component: () => import('@/views/Forgot.vue'),
    meta: { public: true, title: '找回密码' }
  },
  {
    path: '/panel',
    component: () => import('@/views/panel/PanelLayout.vue'),
    meta: { title: '用户面板' },
    children: [
      { path: '', redirect: '/panel/dashboard' },
      { path: 'dashboard', name: 'dashboard', component: () => import('@/views/panel/Dashboard.vue'), meta: { title: '概览' } },
      { path: 'tunnels', name: 'tunnels', component: () => import('@/views/panel/Tunnels.vue'), meta: { title: '隧道管理' } },
      { path: 'stats', name: 'stats', component: () => import('@/views/panel/Stats.vue'), meta: { title: '流量统计' } },
      { path: 'points', name: 'points', component: () => import('@/views/panel/Points.vue'), meta: { title: '积分中心' } },
      { path: 'announcements', name: 'announcements', component: () => import('@/views/panel/Announcements.vue'), meta: { title: '公告' } },
      { path: 'tickets', name: 'tickets', component: () => import('@/views/panel/Tickets.vue'), meta: { title: '工单' } },
      { path: 'profile', name: 'profile', component: () => import('@/views/panel/Profile.vue'), meta: { title: '个人中心' } }
    ]
  },
  {
    path: '/admin/login',
    name: 'adminLogin',
    component: () => import('@/views/admin/AdminLogin.vue'),
    meta: { public: true, title: '管理员登录' }
  },
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { title: '管理后台', admin: true },
    children: [
      { path: '', redirect: '/admin/dashboard' },
      { path: 'dashboard', name: 'adminDashboard', component: () => import('@/views/admin/Dashboard.vue'), meta: { title: '数据看板', admin: true } },
      { path: 'users', name: 'adminUsers', component: () => import('@/views/admin/Users.vue'), meta: { title: '用户管理', admin: true } },
      { path: 'nodes', name: 'adminNodes', component: () => import('@/views/admin/Nodes.vue'), meta: { title: '节点管理', admin: true } },
      { path: 'tunnels', name: 'adminTunnels', component: () => import('@/views/admin/Tunnels.vue'), meta: { title: '隧道管理', admin: true } },
      { path: 'plans', name: 'adminPlans', component: () => import('@/views/admin/Plans.vue'), meta: { title: '套餐配置', admin: true } },
      { path: 'announcements', name: 'adminAnnouncements', component: () => import('@/views/admin/Announcements.vue'), meta: { title: '公告管理', admin: true } },
      { path: 'config', name: 'adminConfig', component: () => import('@/views/admin/Config.vue'), meta: { title: '系统配置', admin: true } },
      { path: 'logs', name: 'adminLogs', component: () => import('@/views/admin/Logs.vue'), meta: { title: '操作日志', admin: true } }
    ]
  },
  { path: '/', redirect: '/panel/dashboard' },
  { path: '/:pathMatch(.*)*', redirect: '/panel/dashboard' }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to) => {
  const userStore = useUserStore()
  if (to.meta.public) {
    if (userStore.isLoggedIn) {
      if (to.path === '/admin/login' && userStore.isAdmin) return '/admin/dashboard'
      if (['/login', '/register', '/forgot'].includes(to.path) && !userStore.isAdmin) return '/panel/dashboard'
    }
    return true
  }
  if (!userStore.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.admin && !userStore.isAdmin) {
    return '/panel/dashboard'
  }
  return true
})

router.afterEach((to) => {
  const title = to.meta.title
  document.title = title ? `${title} - WeaveNet 织网穿透` : 'WeaveNet 织网穿透'
})

export default router
