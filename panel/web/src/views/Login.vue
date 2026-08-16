<template>
  <div class="auth-page">
    <div class="glass-card auth-card">
      <div class="auth-brand">
        <svg width="40" height="40" viewBox="0 0 64 64">
          <defs>
            <linearGradient id="lg-login" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0" stop-color="#0ea5e9" />
              <stop offset="1" stop-color="#06b6d4" />
            </linearGradient>
          </defs>
          <rect width="64" height="64" rx="16" fill="url(#lg-login)" />
          <g fill="none" stroke="#ffffff" stroke-width="4" stroke-linecap="round">
            <circle cx="24" cy="24" r="7" />
            <circle cx="42" cy="42" r="7" />
            <path d="M30 30c4 4 5 7 5 12" />
            <path d="M24 31c-1 6-2 8-6 11" />
          </g>
        </svg>
        <span class="auth-brand-name">WeaveNet 织网穿透</span>
      </div>
      <div class="auth-sub">登录用户面板，畅享高速内网穿透</div>
      <n-form ref="formRef" :model="form" :rules="rules" size="large">
        <n-form-item label="用户名或邮箱" path="account">
          <n-input v-model:value="form.account" placeholder="请输入用户名或邮箱" clearable @keyup.enter="handleLogin">
            <template #prefix>
              <svg-icon name="profile" :size="16" />
            </template>
          </n-input>
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="请输入密码" @keyup.enter="handleLogin">
            <template #prefix>
              <svg-icon name="lock" :size="16" />
            </template>
          </n-input>
        </n-form-item>
        <div class="auth-links">
          <router-link class="auth-link" to="/forgot">忘记密码？</router-link>
          <router-link class="auth-link" to="/register">还没有账号？立即注册</router-link>
        </div>
        <n-button class="btn-grad-orange auth-submit" block size="large" :loading="loading" @click="handleLogin">
          登 录
        </n-button>
      </n-form>
      <div class="auth-foot">登录即代表同意平台《服务条款》与《隐私政策》</div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, NForm, NFormItem, NInput, NButton } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { useUserStore } from '@/store/user'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const userStore = useUserStore()

const formRef = ref(null)
const loading = ref(false)
const form = reactive({ account: '', password: '' })

const rules = {
  account: [{ required: true, message: '请输入用户名或邮箱', trigger: ['input', 'blur'] }],
  password: [{ required: true, message: '请输入密码', trigger: ['input', 'blur'] }]
}

async function handleLogin() {
  try {
    await formRef.value?.validate()
  } catch (e) {
    return
  }
  loading.value = true
  try {
    await userStore.login({
      username: form.account,
      password: form.password
    })
    message.success('登录成功，欢迎回来')
    const redirect = route.query.redirect || (userStore.isAdmin ? '/admin/dashboard' : '/panel/dashboard')
    router.push(String(redirect).startsWith('/') ? redirect : '/panel/dashboard')
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}
</script>
