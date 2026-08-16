<template>
  <div class="auth-page">
    <div class="glass-card auth-card">
      <div class="auth-brand">
        <svg width="40" height="40" viewBox="0 0 64 64">
          <defs>
            <linearGradient id="lg-register" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0" stop-color="#0ea5e9" />
              <stop offset="1" stop-color="#06b6d4" />
            </linearGradient>
          </defs>
          <rect width="64" height="64" rx="16" fill="url(#lg-register)" />
          <g fill="none" stroke="#ffffff" stroke-width="4" stroke-linecap="round">
            <circle cx="24" cy="24" r="7" />
            <circle cx="42" cy="42" r="7" />
            <path d="M30 30c4 4 5 7 5 12" />
            <path d="M24 31c-1 6-2 8-6 11" />
          </g>
        </svg>
        <span class="auth-brand-name">WeaveNet 织网穿透</span>
      </div>
      <div class="auth-sub">注册账号，开启你的穿透之旅</div>

      <template v-if="!needVerify">
        <n-form ref="formRef" :model="form" :rules="rules" size="large">
          <n-form-item label="用户名" path="username">
            <n-input v-model:value="form.username" placeholder="4-20 位字母、数字或下划线" clearable>
              <template #prefix>
                <svg-icon name="profile" :size="16" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="邮箱" path="email">
            <n-input v-model:value="form.email" placeholder="用于接收验证邮件" clearable>
              <template #prefix>
                <svg-icon name="mail" :size="16" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="密码" path="password">
            <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="至少 6 位">
              <template #prefix>
                <svg-icon name="lock" :size="16" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="确认密码" path="confirmPassword">
            <n-input v-model:value="form.confirmPassword" type="password" show-password-on="click" placeholder="再次输入密码" @keyup.enter="handleRegister">
              <template #prefix>
                <svg-icon name="lock" :size="16" />
              </template>
            </n-input>
          </n-form-item>
          <div class="auth-tip">注册后需要前往邮箱查收验证码完成激活，才能正常使用。</div>
          <n-button class="btn-grad-orange auth-submit" block size="large" :loading="loading" @click="handleRegister">
            注 册
          </n-button>
        </n-form>
      </template>

      <template v-else>
        <n-form ref="verifyRef" :model="verify" :rules="verifyRules" size="large">
          <n-form-item label="邮箱" path="email">
            <n-input v-model:value="verify.email" placeholder="请输入注册邮箱" disabled>
              <template #prefix>
                <svg-icon name="mail" :size="16" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="验证码" path="code">
            <n-input v-model:value="verify.code" placeholder="请输入邮件中的验证码" @keyup.enter="handleVerify">
              <template #prefix>
                <svg-icon name="key" :size="16" />
              </template>
            </n-input>
          </n-form-item>
          <div class="auth-tip">验证码已发送至注册邮箱，请查收。验证成功后即可登录使用。</div>
          <n-button class="btn-grad-cyan auth-submit" block size="large" :loading="verifying" @click="handleVerify">
            验 证 并 激 活
          </n-button>
        </n-form>
      </template>

      <div class="auth-foot">
        <router-link class="auth-link" to="/login">已有账号？返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NForm, NFormItem, NInput, NButton } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { register, emailVerify } from '@/api'

const router = useRouter()
const message = useMessage()

const formRef = ref(null)
const verifyRef = ref(null)
const loading = ref(false)
const verifying = ref(false)
const needVerify = ref(false)

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const verify = reactive({
  email: '',
  code: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: ['input', 'blur'] },
    { pattern: /^[a-zA-Z0-9_]{4,20}$/, message: '用户名需为 4-20 位字母、数字或下划线', trigger: ['input', 'blur'] }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: ['input', 'blur'] },
    { type: 'email', message: '邮箱格式不正确', trigger: ['input', 'blur'] }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: ['input', 'blur'] },
    { min: 6, message: '密码至少 6 位', trigger: ['input', 'blur'] }
  ],
  confirmPassword: [
    {
      required: true,
      validator: (rule, value) => value === form.password,
      message: '两次输入的密码不一致',
      trigger: ['input', 'blur']
    }
  ]
}

const verifyRules = {
  email: [{ required: true, message: '请输入注册邮箱', trigger: ['input', 'blur'] }],
  code: [{ required: true, message: '请输入验证码', trigger: ['input', 'blur'] }]
}

async function handleRegister() {
  try {
    await formRef.value?.validate()
  } catch (e) {
    return
  }
  loading.value = true
  try {
    await register({
      username: form.username,
      email: form.email,
      password: form.password
    })
    message.success('注册成功，验证码已发送至邮箱')
    verify.email = form.email
    needVerify.value = true
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

async function handleVerify() {
  try {
    await verifyRef.value?.validate()
  } catch (e) {
    return
  }
  verifying.value = true
  try {
    await emailVerify({
      email: verify.email,
      code: verify.code,
      purpose: 'register'
    })
    message.success('邮箱验证成功，请登录')
    router.push('/login')
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    verifying.value = false
  }
}
</script>
