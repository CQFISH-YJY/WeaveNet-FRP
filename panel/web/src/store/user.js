import { defineStore } from 'pinia'
import { login as apiLogin, logout as apiLogout, getProfile } from '@/api'

const TOKEN_KEY = 'weavenet_token'
const USER_KEY = 'weavenet_user'

function readUser() {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || 'null')
  } catch (e) {
    return null
  }
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    userInfo: readUser()
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.role === 'admin',
    displayName: (state) => state.userInfo?.username || state.userInfo?.nickname || '用户'
  },
  actions: {
    async login(payload) {
      const data = await apiLogin(payload)
      this.token = data.token || data.access_token || ''
      this.userInfo = data.user || data
      localStorage.setItem(TOKEN_KEY, this.token)
      localStorage.setItem(USER_KEY, JSON.stringify(this.userInfo))
      return data
    },
    async fetchProfile() {
      const data = await getProfile()
      this.userInfo = data
      localStorage.setItem(USER_KEY, JSON.stringify(this.userInfo))
      return data
    },
    async logout() {
      try {
        await apiLogout()
      } catch (e) {
        // 忽略登出接口异常，本地一定清理
      }
      this.clearAuth()
    },
    clearAuth() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    }
  }
})
