<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>系统配置</h2>
        <div class="sub">管理签到规则、积分兑换与邮件服务配置</div>
      </div>
      <n-button class="btn-grad-cyan" :loading="saving" @click="saveAll">
        <template #icon><svg-icon name="check" :size="15" /></template>
        保存全部
      </n-button>
    </div>

    <div v-for="group in groups" :key="group.key" class="glass-card group-card">
      <div class="group-head">
        <span class="group-title">
          <svg-icon :name="group.icon" :size="16" class="group-icon" />
          {{ group.label }}
        </span>
        <span class="group-desc">{{ group.desc }}</span>
      </div>
      <div class="config-list">
        <div v-for="item in group.items" :key="item.key" class="config-row">
          <div class="config-label">
            <div>{{ labelMap[item.key] || item.key }}</div>
            <div class="config-key">{{ item.key }}</div>
          </div>
          <div class="config-input">
            <n-input-number
              v-if="item.type === 'number'"
              v-model:value="item.value"
              style="width: 280px"
              :min="0"
            />
            <n-switch v-else-if="item.type === 'boolean'" v-model:value="item.value" />
            <n-input v-else v-model:value="item.value" style="width: 320px" />
          </div>
        </div>
        <div v-if="!group.items.length" class="empty-tip">暂无配置项</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useMessage, NButton, NInput, NInputNumber, NSwitch } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { adminGetConfig, adminUpdateConfig } from '@/api'

const message = useMessage()
const saving = ref(false)
const configMap = reactive({})

const labelMap = {
  signin_points: '每日签到积分',
  continuous_days: '连续签到天数',
  continuous_bonus: '连续签到奖励',
  exchange_cost: '兑换所需积分',
  exchange_plan: '兑换套餐名称',
  smtp_host: 'SMTP 服务器',
  smtp_port: 'SMTP 端口',
  smtp_user: 'SMTP 账号',
  smtp_pass: 'SMTP 密码',
  smtp_from: '发件人地址',
  smtp_tls: '启用 TLS'
}

function classify(key) {
  if (key.includes('signin') || key.includes('continuous')) return 'signin'
  if (key.includes('exchange')) return 'exchange'
  if (key.includes('smtp')) return 'smtp'
  return 'other'
}

const flatItems = computed(() =>
  Object.entries(configMap).map(([key, value]) => ({
    key,
    value,
    type: typeof value === 'number' ? 'number' : typeof value === 'boolean' ? 'boolean' : 'string'
  }))
)

const groups = computed(() => {
  const defs = [
    { key: 'signin', label: '签到规则', icon: 'check', desc: '每日签到积分与连续签到奖励', match: (k) => k.includes('signin') || k.includes('continuous') },
    { key: 'exchange', label: '积分兑换', icon: 'plan', desc: '积分兑换会员所需积分', match: (k) => k.includes('exchange') },
    { key: 'smtp', label: '邮件 SMTP', icon: 'mail', desc: '发送验证码邮件的 SMTP 服务', match: (k) => k.includes('smtp') },
    { key: 'other', label: '其他配置', icon: 'config', desc: '其余系统配置项', match: () => true }
  ]
  return defs
    .map((d) => ({
      ...d,
      items: flatItems.value.filter((i) => d.match(i.key))
    }))
    .filter((g) => g.items.length)
})

async function loadData() {
  try {
    const data = (await adminGetConfig()) || {}
    Object.keys(configMap).forEach((k) => delete configMap[k])
    if (Array.isArray(data)) {
      data.forEach((item) => {
        configMap[item.key] = normalize(item.value)
      })
    } else {
      Object.entries(data).forEach(([key, value]) => {
        configMap[key] = normalize(value)
      })
    }
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function normalize(value) {
  if (value === 'true' || value === 'false') return value === 'true'
  const num = Number(value)
  if (value !== '' && !Number.isNaN(num) && typeof value === 'string' && value.trim() !== '') return num
  return value
}

async function saveAll() {
  saving.value = true
  try {
    for (const item of flatItems.value) {
      await adminUpdateConfig({ key: item.key, value: item.value })
    }
    message.success('全部配置已保存')
    await loadData()
  } catch (e) {
    message.error('部分配置保存失败，请检查后重试')
  } finally {
    saving.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.group-card {
  padding: 20px;
  margin-bottom: 16px;
}

.group-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

.group-icon {
  color: #0ea5e9;
}

.group-desc {
  font-size: 12px;
  color: #94a3b8;
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.config-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.config-label {
  font-size: 13px;
  color: #334155;
}

.config-key {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 2px;
  font-family: Consolas, 'Courier New', monospace;
}

.config-input {
  flex-shrink: 0;
}
</style>
