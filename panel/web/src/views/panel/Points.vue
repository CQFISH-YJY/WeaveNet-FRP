<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>积分中心</h2>
        <div class="sub">每日签到赚积分，兑换会员权益</div>
      </div>
    </div>

    <div class="points-grid">
      <div class="glass-card points-card main">
        <div class="points-amount">
          <div class="stat-label">当前积分余额</div>
          <div class="points-num">{{ points }}</div>
        </div>
        <div class="points-actions">
          <n-button
            v-if="!signinStatus.signed_in"
            class="btn-grad-orange"
            size="large"
            :loading="signing"
            @click="handleSignin"
          >
            <template #icon><svg-icon name="check" :size="16" /></template>
            每日签到
          </n-button>
          <div v-else class="signed-tip">
            今日已签到，连续 {{ signinStatus.continuous_days || 0 }} 天
          </div>
          <n-button class="btn-grad-cyan" :loading="exchanging" @click="handleExchange">
            <template #icon><svg-icon name="plan" :size="16" /></template>
            兑换普通会员
          </n-button>
        </div>
      </div>

      <div class="glass-card points-card rules">
        <div class="card-head">
          <span class="card-head-title">
            <svg-icon name="info" :size="16" class="head-icon" />
            积分规则
          </span>
        </div>
        <ul class="rules-list">
          <li v-for="(r, i) in ruleItems" :key="i">{{ r }}</li>
        </ul>
      </div>
    </div>

    <div class="glass-card logs-card">
      <div class="card-head">
        <span class="card-head-title">
          <svg-icon name="logs" :size="16" class="head-icon" />
          积分流水
        </span>
      </div>
      <n-table :bordered="false" size="small">
        <thead>
          <tr>
            <th>时间</th>
            <th>变动</th>
            <th>说明</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id">
            <td>{{ formatDate(log.created_at) }}</td>
            <td>
              <span :class="log.change > 0 ? 'change-in' : 'change-out'">
                {{ log.change > 0 ? '+' : '' }}{{ log.change }}
              </span>
            </td>
            <td>{{ log.reason || '-' }}</td>
          </tr>
          <tr v-if="!logs.length">
            <td colspan="3">
              <div class="empty-tip">暂无积分流水</div>
            </td>
          </tr>
        </tbody>
      </n-table>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useMessage, useDialog, NButton, NTable } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { getPointsLogs, getPointsRules, getSigninStatus, signin, exchangePoints } from '@/api'
import { formatDate } from '@/utils/format'
import { useUserStore } from '@/store/user'

const message = useMessage()
const dialog = useDialog()
const userStore = useUserStore()

const logs = ref([])
const rules = ref({})
const signinStatus = reactive({ signed_in: false, continuous_days: 0 })
const signing = ref(false)
const exchanging = ref(false)

const points = computed(() => userStore.userInfo?.points ?? 0)

const ruleItems = computed(() => {
  const r = rules.value
  const items = []
  if (r && typeof r === 'object') {
    if (r.signin_points) items.push(`每日签到可得 ${r.signin_points} 积分`)
    if (r.continuous_days && r.continuous_bonus) {
      items.push(`连续签到 ${r.continuous_days} 天额外奖励 ${r.continuous_bonus} 积分`)
    }
    if (r.exchange_cost) items.push(`兑换 1 个月普通会员需 ${r.exchange_cost} 积分`)
    if (r.exchange_plan) items.push(`积分可兑换套餐：${r.exchange_plan}`)
  }
  if (!items.length) {
    items.push('每日签到得 10 积分，连续 7 天额外 +30 积分')
    items.push('300 积分可兑换 1 个月普通会员')
    items.push('积分规则以系统配置为准')
  }
  return items
})

async function loadData() {
  try {
    const [logList, rule, status] = await Promise.allSettled([
      getPointsLogs(),
      getPointsRules(),
      getSigninStatus()
    ])
    if (logList.status === 'fulfilled') logs.value = Array.isArray(logList.value) ? logList.value : []
    if (rule.status === 'fulfilled') rules.value = rule.value || {}
    if (status.status === 'fulfilled') Object.assign(signinStatus, status.value || {})
    if (status.status === 'fulfilled' || userStore.userInfo) {
      try {
        await userStore.fetchProfile()
      } catch (e) {
        // 忽略
      }
    }
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function handleSignin() {
  signing.value = true
  try {
    const data = await signin()
    message.success(`签到成功${data?.points ? `，获得 ${data.points} 积分` : ''}`)
    signinStatus.signed_in = true
    signinStatus.continuous_days = data?.continuous_days ?? signinStatus.continuous_days
    await userStore.fetchProfile()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    signing.value = false
  }
}

function handleExchange() {
  dialog.warning({
    title: '兑换普通会员',
    content: `将使用 ${rules.value.exchange_cost || 300} 积分兑换 1 个月普通会员，确定继续吗？`,
    positiveText: '确认兑换',
    negativeText: '取消',
    onPositiveClick: async () => {
      exchanging.value = true
      try {
        await exchangePoints({})
        message.success('兑换成功，普通会员已到账')
        await userStore.fetchProfile()
      } catch (e) {
        // 错误已由拦截器提示
      } finally {
        exchanging.value = false
      }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.points-grid {
  display: grid;
  grid-template-columns: 1.3fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.points-card {
  padding: 22px;
}

.points-num {
  font-size: 42px;
  font-weight: 800;
  line-height: 1.2;
  background: linear-gradient(135deg, #f59e0b, #f97316);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  margin: 6px 0 18px;
}

.points-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.signed-tip {
  font-size: 13px;
  color: #059669;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
}

.card-head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 12px;
}

.head-icon {
  color: #0ea5e9;
}

.rules-list {
  margin: 0;
  padding-left: 18px;
  font-size: 13px;
  color: #475569;
  line-height: 2;
}

.logs-card {
  padding: 20px;
}

.change-in {
  color: #059669;
  font-weight: 600;
}

.change-out {
  color: #dc2626;
  font-weight: 600;
}

@media (max-width: 900px) {
  .points-grid {
    grid-template-columns: 1fr;
  }
}
</style>
