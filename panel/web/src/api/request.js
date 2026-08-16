import axios from 'axios'
import { createDiscreteApi } from 'naive-ui'

let msg = null

function getMessage() {
  if (!msg) {
    const api = createDiscreteApi(['message'])
    msg = api.message
  }
  return msg
}

const request = axios.create({
  baseURL: '/api',
  timeout: 15000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('weavenet_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res && res.code === 0) {
      return res.data
    }
    const errMsg = (res && res.message) || '请求失败'
    getMessage().error(errMsg)
    return Promise.reject(new Error(errMsg))
  },
  (error) => {
    const status = error.response?.status
    const res = error.response?.data
    if (status === 401) {
      localStorage.removeItem('weavenet_token')
      localStorage.removeItem('weavenet_user')
      const current = window.location.hash.replace(/^#/, '') || '/panel/dashboard'
      if (!current.startsWith('/login') && !current.startsWith('/admin/login')) {
        getMessage().error('登录已过期，请重新登录')
        setTimeout(() => {
          window.location.hash = `#/login?redirect=${encodeURIComponent(current)}`
        }, 300)
      }
    } else {
      getMessage().error(res?.message || error.message || '网络异常，请稍后重试')
    }
    return Promise.reject(error)
  }
)

export default request
