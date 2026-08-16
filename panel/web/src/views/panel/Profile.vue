<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>个人中心</h2>
        <div class="sub">管理你的账号信息与安全设置</div>
      </div>
    </div>

    <div class="grid-2">
      <div class="glass-card info-card">
        <div class="card-head">
          <span class="card-head-title">
            <svg-icon name="profile" :size="17" class="head-icon" />
            账号信息
          </span>
        </div>
        <div class="profile-top">
          <div class="avatar-lg">{{ avatarText }}</div>
          <div>
            <div class="name-lg">{{ userStore.displayName }}</div>
            <div class="plan-chip">{{ userInfo.plan_name || '免费版' }}</div>
          </div>
        </div>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">用户名</span>
            <span class="info-value">{{ userInfo.username || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">邮箱</span>
            <span class="info-value">{{ userInfo.email || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">邮箱验证</span>
            <span class="info-value">
              <status-badge :type="emailVerified ? 'success' : 'warning'" :text="emailVerified ? '已验证' : '未验证'" />
            </span>
          </div>
          <div class="info-item">
            <span class="info-label">积分</span>
            <span class="info-value warn">{{ userInfo.points ?? 0 }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">套餐到期</span>
            <span class="info-value">{{ planExpireText(userInfo.plan_expires_at) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">注册时间</span>
            <span class="info-value">{{ formatDate(userInfo.created_at, 'YYYY-MM-DD') }}</span>
          </div>
        </div>
      </div>

      <div class="glass-card info-card">
        <div class="card-head">
          <span class="card-head-title">
            <svg-icon name="lock" :size="17" class="head-icon" />
            修改密码
          </span>
        </div>
        <n-form ref="pwdRef" :model="pwdForm" :rules="pwdRules" label-placement="top">
          <n-form-item label="当前密码" path="old_password">
            <n-input v-model:value="pwdForm.old_password" type="password" show-password-on="click" placeholder="请输入当前密码">
              <template #prefix><svg-icon name="lock" :size="15" /></template>
            </n-input>
          </n-form-item>
          <n-form-item label="新密码" path="new_password">
            <n-input v-model:value="pwdForm.new_password" type="password" show-password-on="click" placeholder="至少 6 位">
              <template #prefix><svg-icon name="key" :size="15" /></template>
            </n-input>
          </n-form-item>
          <n-form-item label="确认新密码" path="confirm">
            <n-input v-model:value="pwdForm.confirm" type="password" show-password-on="click" placeholder="再次输入新密码">
              <template #prefix><svg-icon name="key" :size="15" /></template>
            </n-input>
          </n-form-item>
          <n-button class="btn-grad-cyan submit-btn" block :loading="changing" @click="handleChangePwd">保存新密码</n-button>
        </n-form>
        <div class="logout-area">
          <n-button type="error" ghost block @click="handleLogout">
            <template #icon><svg-icon name="logout" :size="15" /></template>
            退出登录
          </n-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, useDialog, NButton, NForm, NFormItem, NInput } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useUserStore } from '@/store/user'
import { updatePassword } from '@/api'
import { formatDate, planExpireText } from '@/utils/format'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const userStore = useUserStore()

const userInfo = computed(() => userStore.userInfo || {})
const avatarText = computed(() => (userStore.displayName || 'U').slice(0, 1).toUpperCase())
const emailVerified = computed(() => {
  const v = userInfo.value.email_verified
  return v === true || v === 1 || v === '1' || v === 'true'
})

const pwdRef = ref(null)
const changing = ref(false)
const pwdForm = reactive({ old_password: '', new_password: '', confirm: '' })

const pwdRules = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: ['input', 'blur'] }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: ['input', 'blur'] },
    { min: 6, message: '密码至少 6 位', trigger: ['input', 'blur'] }
  ],
  confirm: [
    {
      required: true,
      validator: (rule, value) => value === pwdForm.new_password,
      message: '两次输入的密码不一致',
      trigger: ['input', 'blur']
    }
  ]
}

async function handleChangePwd() {
  try {
    await pwdRef.value?.validate()
  } catch (e) {
    return
  }
  changing.value = true
  try {
    await updatePassword({
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password
    })
    message.success('密码修改成功')
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm = ''
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    changing.value = false
  }
}

function handleLogout() {
  dialog.warning({
    title: '退出登录',
    content: '确定要退出当前账号吗？',
    positiveText: '退出',
    negativeText: '取消',
    onPositiveClick: async () => {
      await userStore.logout()
      message.success('已安全退出')
      router.push('/login')
    }
  })
}

onMounted(() => {
  userStore.fetchProfile().catch(() => {})
})
</script>

<style scoped>
.grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.info-card {
  padding: 22px;
}

.card-head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 16px;
}

.head-icon {
  color: #0ea5e9;
}

.profile-top {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 18px;
}

.avatar-lg {
  width: 60px;
  height: 60px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
}

.name-lg {
  font-size: 18px;
  font-weight: 600;
  color: #0f172a;
}

.plan-chip {
  display: inline-block;
  margin-top: 6px;
  padding: 2px 12px;
  border-radius: 999px;
  font-size: 12px;
  color: #b45309;
  background: rgba(245, 158, 11, 0.14);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.info-item {
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.info-label {
  display: block;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 4px;
}

.info-value {
  font-size: 13px;
  color: #334155;
  word-break: break-all;
}

.info-value.warn {
  color: #f59e0b;
  font-weight: 600;
}

.submit-btn {
  margin-top: 4px;
}

.logout-area {
  margin-top: 20px;
}

@media (max-width: 900px) {
  .grid-2 {
    grid-template-columns: 1fr;
  }
}
</style>
