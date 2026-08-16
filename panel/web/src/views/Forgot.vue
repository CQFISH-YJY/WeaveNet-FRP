<template>
  <div class="auth-page">
    <div class="glass-card auth-card">
      <div class="auth-brand">
        <svg width="40" height="40" viewBox="0 0 64 64">
          <defs>
            <linearGradient id="lg-forgot" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0" stop-color="#0ea5e9" />
              <stop offset="1" stop-color="#06b6d4" />
            </linearGradient>
          </defs>
          <rect width="64" height="64" rx="16" fill="url(#lg-forgot)" />
          <g fill="none" stroke="#ffffff" stroke-width="4" stroke-linecap="round">
            <circle cx="24" cy="24" r="7" />
            <circle cx="42" cy="42" r="7" />
            <path d="M30 30c4 4 5 7 5 12" />
            <path d="M24 31c-1 6-2 8-6 11" />
          </g>
        </svg>
        <span class="auth-brand-name">WeaveNet 织网穿透</span>
      </div>
      <div class="auth-sub">找回密码，验证邮箱后重置</div>
      <n-form ref="formRef" :model="form" :rules="rules" size="large">
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="请输入注册邮箱" clearable>
            <template #prefix>
              <svg-icon name="mail" :size="16" />
            </template>
          </n-input>
        </n-form-item>
        <n-form-item label="验证码" path="code">
          <div style="display: flex; gap: 8px; width: 100%">
            <n-input v-model:value="form.code" placeholder="邮件中的验证码">
              <template #prefix>
                <svg-icon name="key" :size="16" />
              </template>
            </n-input>
            <n-button :loading="sending" @click="handleSendCode">发送验证码</n-button>
          </div>
        </n-form-item>
        <n-form-item label="新密码" path="password">
          <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="设置新密码，至少 6 位">
            <template #prefix>
              <svg-icon name="lock" :size="16" />
            </template>
          </n-input>
        </n-form-item>
        <n-form-item label="确认新密码" path="confirmPassword">
          <n-input v-model:value="form.confirmPassword" type="password" show-password-on="click" placeholder="再次输入新密码" @keyup.enter="handleReset">
            <template #prefix>
              <svg-icon name="lock" :size="16" />
            </template>
          </n-input>
        </n-form-item>
        <div class="auth-tip">点击发送验证码后，系统会向邮箱发送一封重置邮件，验证通过后即可重置密码。</div>
        <n-button class="btn-grad-orange auth-submit" block size="large" :loading="loading" @click="handleReset">
          重 置 密 码
        </n-button>
      </n-form>
      <div class="auth-foot">
        <router-link class="auth-link" to="/login">想起密码了？返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NForm, NFormItem, NInput, NButton } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { forgotPassword, emailVerify } from '@/api'

const router = useRouter()
const message = useMessage()

const formRef = ref(null)
const loading = ref(false)
const sending = ref(false)

const form = reactive({
  email: '',
  code: '',
  password: '',
  confirmPassword: ''
})

const rules = {
  email: [
    { required: true, message: '请输入注册邮箱', trigger: ['input', 'blur'] },
    { type: 'email', message: '邮箱格式不正确', trigger: ['input', 'blur'] }
  ],
  code: [{ required: true, message: '请输入验证码', trigger: ['input', 'blur'] }],
  password: [
    { required: true, message: '请输入新密码', trigger: ['input', 'blur'] },
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

async function handleSendCode() {
  if (!form.email) {
    message.warning('请先填写邮箱')
    return
  }
  sending.value = true
  try {
    await forgotPassword({ email: form.email })
    message.success('重置邮件已发送，请查收验证码')
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    sending.value = false
  }
}

async function handleReset() {
  try {
    await formRef.value?.validate()
  } catch (e) {
    return
  }
  loading.value = true
  try {
    await emailVerify({
      email: form.email,
      code: form.code,
      purpose: 'reset',
      new_password: form.password
    })
    message.success('密码重置成功，请使用新密码登录')
    router.push('/login')
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}
</script>
