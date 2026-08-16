<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>你好，{{ userStore.displayName }}</h2>
        <div class="sub">欢迎回到 WeaveNet 织网穿透控制台</div>
      </div>
      <n-button class="btn-grad-cyan" @click="router.push('/panel/tunnels')">
        <template #icon>
          <svg-icon name="plus" :size="16" />
        </template>
        创建隧道
      </n-button>
    </div>

    <div class="stat-grid">
      <div class="glass-card stat-card">
        <div class="stat-icon orange">
          <svg-icon name="plan" :size="22" />
        </div>
        <div class="stat-info">
          <div class="stat-label">当前套餐</div>
          <div class="stat-value plan-name">{{ profile.plan_name || '免费版' }}</div>
          <div class="stat-meta">
            限速 {{ quota.speed_limit_mbps ?? '-' }} Mbps
            <span class="dot">/</span>
            隧道 {{ quota.tunnel_limit ?? '-' }} 条
          </div>
          <div class="stat-meta">到期时间：{{ planExpireText(profile.plan_expires_at) }}</div>
        </div>
      </div>

      <div class="glass-card stat-card">
        <div class="stat-icon cyan">
          <svg-icon name="stats" :size="22" />
        </div>
        <div class="stat-info">
          <div class="stat-label">今日流量</div>
          <div class="stat-value">{{ formatBytes(overview.today_total) }}</div>
          <div class="stat-meta">上行 {{ formatBytes(overview.today_in) }}</div>
          <div class="stat-meta">下行 {{ formatBytes(overview.today_out) }}</div>
        </div>
      </div>

      <div class="glass-card stat-card">
        <div class="stat-icon green">
          <svg-icon name="wifi" :size="22" />
        </div>
        <div class="stat-info">
          <div class="stat-label">在线隧道</div>
          <div class="stat-value">{{ overview.online_tunnels ?? 0 }}</div>
          <div class="stat-meta">共 {{ overview.total_tunnels ?? 0 }} 条隧道</div>
          <n-button text size="small" class="link-btn" @click="router.push('/panel/tunnels')">
            前往隧道管理
            <svg-icon name="arrowRight" :size="14" />
          </n-button>
        </div>
      </div>

      <div class="glass-card stat-card">
        <div class="stat-icon warm">
          <svg-icon name="points" :size="22" />
        </div>
        <div class="stat-info">
          <div class="stat-label">我的积分</div>
          <div class="stat-value">{{ profile.points ?? 0 }}</div>
          <div class="stat-meta">
            <template v-if="signinStatus.signed_in">
              今日已签到，连续 {{ signinStatus.continuous_days || 0 }} 天
            </template>
            <template v-else>今日还未签到，快来签到吧</template>
          </div>
          <n-button
            v-if="!signinStatus.signed_in"
            class="btn-grad-orange"
            size="small"
            :loading="signing"
            @click="handleSignin"
          >
            立即签到
          </n-button>
          <n-button v-else text size="small" class="link-btn" @click="router.push('/panel/points')">
            前往积分中心
            <svg-icon name="arrowRight" :size="14" />
          </n-button>
        </div>
      </div>
    </div>

    <div class="lower-grid">
      <div class="glass-card lower-card">
        <div class="lower-title">
          <span>最新公告</span>
          <n-button text size="small" class="link-btn" @click="router.push('/panel/announcements')">
            查看全部
            <svg-icon name="arrowRight" :size="14" />
          </n-button>
        </div>
        <div v-if="announcements.length" class="ann-list">
          <div v-for="item in announcements.slice(0, 4)" :key="item.id" class="ann-item" @click="router.push('/panel/announcements')">
            <span class="ann-title">{{ item.title }}</span>
            <span class="ann-meta">{{ item.author }} · {{ formatDate(item.created_at, 'YYYY-MM-DD') }}</span>
          </div>
        </div>
        <div v-else class="empty-tip">暂无公告</div>
      </div>

      <div class="glass-card lower-card">
        <div class="lower-title">
          <span>快速入口</span>
        </div>
        <div class="quick-grid">
          <div class="quick-item" @click="router.push('/panel/tunnels')">
            <span class="quick-icon cyan"><svg-icon name="tunnel" :size="20" /></span>
            <span>隧道管理</span>
          </div>
          <div class="quick-item" @click="router.push('/panel/stats')">
            <span class="quick-icon green"><svg-icon name="stats" :size="20" /></span>
            <span>流量统计</span>
          </div>
          <div class="quick-item" @click="router.push('/panel/tickets')">
            <span class="quick-icon orange"><svg-icon name="ticket" :size="20" /></span>
            <span>提交工单</span>
          </div>
          <div class="quick-item" @click="router.push('/panel/profile')">
            <span class="quick-icon warm"><svg-icon name="profile" :size="20" /></span>
            <span>个人中心</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NButton } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { useUserStore } from '@/store/user'
import { getQuota, getStatsOverview, getSigninStatus, getAnnouncements, signin } from '@/api'
import { formatBytes, formatDate, planExpireText } from '@/utils/format'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const profile = reactive({})
const quota = reactive({})
const overview = reactive({})
const signinStatus = reactive({ signed_in: false, continuous_days: 0 })
const announcements = ref([])
const signing = ref(false)

async function loadAll() {
  try {
    const [p, q, o, s, anns] = await Promise.allSettled([
      userStore.fetchProfile(),
      getQuota(),
      getStatsOverview(),
      getSigninStatus(),
      getAnnouncements()
    ])
    if (p.status === 'fulfilled') Object.assign(profile, p.value || {})
    if (q.status === 'fulfilled') Object.assign(quota, q.value || {})
    if (o.status === 'fulfilled') Object.assign(overview, o.value || {})
    if (s.status === 'fulfilled') Object.assign(signinStatus, s.value || {})
    if (anns.status === 'fulfilled') announcements.value = Array.isArray(anns.value) ? anns.value : []
  } catch (e) {
    // 单个接口失败不影响整体展示
  }
}

async function handleSignin() {
  signing.value = true
  try {
    await signin()
    message.success('签到成功，积分已到账')
    signinStatus.signed_in = true
    await loadAll()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    signing.value = false
  }
}

onMounted(loadAll)
</script>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  display: flex;
  gap: 14px;
  padding: 20px;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.orange {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.18), rgba(249, 115, 22, 0.18));
  color: #f97316;
}

.stat-icon.cyan {
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.16), rgba(6, 182, 212, 0.16));
  color: #0ea5e9;
}

.stat-icon.green {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.16), rgba(5, 150, 105, 0.16));
  color: #10b981;
}

.stat-icon.warm {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.16), rgba(251, 191, 36, 0.18));
  color: #f59e0b;
}

.stat-info {
  min-width: 0;
  flex: 1;
}

.plan-name {
  font-size: 22px;
}

.stat-meta {
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
}

.stat-meta .dot {
  margin: 0 4px;
  color: #cbd5e1;
}

.link-btn {
  color: #0ea5e9 !important;
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-top: 6px;
}

.lower-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.lower-card {
  padding: 20px;
}

.lower-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 12px;
}

.ann-list {
  display: flex;
  flex-direction: column;
}

.ann-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 8px;
  border-radius: 10px;
  cursor: pointer;
}

.ann-item:hover {
  background: rgba(14, 165, 233, 0.06);
}

.ann-title {
  font-size: 13px;
  color: #334155;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 60%;
}

.ann-meta {
  font-size: 12px;
  color: #94a3b8;
  flex-shrink: 0;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.quick-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px;
  border-radius: 14px;
  cursor: pointer;
  font-size: 13px;
  color: #475569;
  background: rgba(255, 255, 255, 0.45);
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.quick-item:hover {
  background: rgba(14, 165, 233, 0.08);
  color: #0284c7;
}

.quick-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.quick-icon.cyan {
  background: rgba(14, 165, 233, 0.14);
  color: #0ea5e9;
}

.quick-icon.green {
  background: rgba(16, 185, 129, 0.14);
  color: #10b981;
}

.quick-icon.orange {
  background: rgba(249, 115, 22, 0.14);
  color: #f97316;
}

.quick-icon.warm {
  background: rgba(245, 158, 11, 0.16);
  color: #f59e0b;
}

@media (max-width: 1100px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .lower-grid {
    grid-template-columns: 1fr;
  }
}
</style>
