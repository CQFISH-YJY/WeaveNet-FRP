import request from './request'

// ---------- 认证 ----------
export const register = (data) => request.post('/auth/register', data)
export const emailVerify = (data) => request.post('/auth/email-verify', data)
export const forgotPassword = (data) => request.post('/auth/forgot-password', data)
export const login = (data) => request.post('/auth/login', data)
export const logout = () => request.post('/auth/logout')

// ---------- 用户 ----------
export const getProfile = () => request.get('/user/profile')
export const updatePassword = (data) => request.put('/user/password', data)
export const getQuota = () => request.get('/user/quota')

// ---------- 节点 ----------
export const getNodes = () => request.get('/nodes')

// ---------- 隧道 ----------
export const getTunnels = () => request.get('/tunnels')
export const createTunnel = (data) => request.post('/tunnels', data)
export const getTunnelDetail = (id) => request.get(`/tunnels/${id}`)
export const updateTunnel = (id, data) => request.put(`/tunnels/${id}`, data)
export const deleteTunnel = (id) => request.delete(`/tunnels/${id}`)
export const startTunnel = (id) => request.post(`/tunnels/${id}/start`)
export const stopTunnel = (id) => request.post(`/tunnels/${id}/stop`)
export const getTunnelConfig = (id) => request.post(`/tunnels/${id}/config`)

// ---------- 统计 ----------
export const getTrafficStats = (params) => request.get('/stats/traffic', { params })
export const getStatsOverview = () => request.get('/stats/overview')

// ---------- 签到 ----------
export const signin = () => request.post('/signin')
export const getSigninStatus = () => request.get('/signin/status')

// ---------- 积分 ----------
export const getPointsLogs = (params) => request.get('/points/logs', { params })
export const exchangePoints = (data) => request.post('/points/exchange', data)
export const getPointsRules = () => request.get('/points/rules')

// ---------- 公告 ----------
export const getAnnouncements = () => request.get('/announcements')
export const getAnnouncementDetail = (id) => request.get(`/announcements/${id}`)

// ---------- 工单 ----------
export const getTickets = () => request.get('/tickets')
export const getTicketDetail = (id) => request.get(`/tickets/${id}`)
export const createTicket = (data) => request.post('/tickets', data)
export const replyTicket = (id, data) => request.post(`/tickets/${id}/reply`, data)
export const closeTicket = (id) => request.post(`/tickets/${id}/close`)

// ---------- 管理后台：用户 ----------
export const adminGetUsers = (params) => request.get('/admin/users', { params })
export const adminBanUser = (id) => request.post(`/admin/users/${id}/ban`)
export const adminUnbanUser = (id) => request.post(`/admin/users/${id}/unban`)
export const adminResetUserPassword = (id) => request.post(`/admin/users/${id}/reset-password`)
export const adminSetUserPlan = (id, data) => request.put(`/admin/users/${id}/plan`, data)

// ---------- 管理后台：节点 ----------
export const adminGetNodes = () => request.get('/admin/nodes')
export const adminCreateNode = (data) => request.post('/admin/nodes', data)
export const adminUpdateNode = (id, data) => request.put(`/admin/nodes/${id}`, data)
export const adminDeleteNode = (id) => request.delete(`/admin/nodes/${id}`)
export const adminStartNode = (id) => request.post(`/admin/nodes/${id}/start`)
export const adminStopNode = (id) => request.post(`/admin/nodes/${id}/stop`)
export const adminSetNodeSpeed = (id, data) => request.put(`/admin/nodes/${id}/speed`, data)

// ---------- 管理后台：隧道 ----------
export const adminGetTunnels = (params) => request.get('/admin/tunnels', { params })
export const adminOfflineTunnel = (id) => request.post(`/admin/tunnels/${id}/offline`)

// ---------- 管理后台：套餐 ----------
export const adminGetPlans = () => request.get('/admin/plans')
export const adminUpdatePlan = (id, data) => request.put(`/admin/plans/${id}`, data)

// ---------- 管理后台：公告 ----------
export const adminGetAnnouncements = () => request.get('/admin/announcements')
export const adminCreateAnnouncement = (data) => request.post('/admin/announcements', data)
export const adminUpdateAnnouncement = (id, data) => request.put(`/admin/announcements/${id}`, data)
export const adminOfflineAnnouncement = (id) => request.post(`/admin/announcements/${id}/offline`)

// ---------- 管理后台：系统配置 ----------
export const adminGetConfig = () => request.get('/admin/config')
export const adminUpdateConfig = (data) => request.put('/admin/config', data)

// ---------- 管理后台：日志与看板 ----------
export const adminGetOperationLogs = (params) => request.get('/admin/logs/operation', { params })
export const adminGetDashboard = () => request.get('/admin/dashboard')
