<template>
  <div class="home">
    <!-- 未登录提示 -->
    <div v-if="!store.loggedIn" class="login-tip card">
      <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21c1.5-3.5 4.5-5 8-5s6.5 1.5 8 5" />
      </svg>
      <span>登录后可查看流量统计、签到领积分并管理隧道</span>
      <button class="btn btn-primary btn-sm" @click="$emit('open-login')">立即登录</button>
    </div>

    <!-- 用户信息卡 -->
    <div class="card home-hero">
      <div class="hero-left">
        <h2>你好，{{ store.loggedIn ? (store.user && store.user.username) : '访客' }}</h2>
        <p class="hero-desc">
          {{ store.loggedIn
            ? `当前套餐：${(store.user && store.user.plan_name) || '免费版'}`
            : '未登录状态，可浏览各项功能，登录后开启完整能力' }}
        </p>
        <div class="hero-stats" v-if="store.loggedIn">
          <div class="hero-stat">
            <span class="stat-num">{{ quota.tunnel_count ?? '-' }}</span>
            <span class="stat-label">隧道</span>
          </div>
          <div class="hero-stat">
            <span class="stat-num">{{ (quota.plan && quota.plan.speed_limit_mbps) || '-' }}</span>
            <span class="stat-label">限速 Mbps</span>
          </div>
          <div class="hero-stat">
            <span class="stat-num">{{ (store.user && store.user.points) ?? '-' }}</span>
            <span class="stat-label">积分</span>
          </div>
        </div>
        <div v-else class="hero-stats">
          <div class="hero-stat">
            <span class="stat-num">{{ store.tunnels.length }}</span>
            <span class="stat-label">已浏览隧道</span>
          </div>
        </div>
      </div>
      <div class="hero-right">
        <button v-if="!store.loggedIn" class="btn btn-primary" @click="$emit('open-login')">登录账号</button>
        <button v-else class="btn btn-ghost" @click="doSignin" :disabled="signing || signedToday">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 3l7 3v5c0 4.6-3 8.1-7 10-4-1.9-7-5.4-7-10V6l7-3z" />
            <path d="M9 12l2 2 4-4" />
          </svg>
          {{ signedToday ? '今日已签到' : (signing ? '签到中...' : '每日签到') }}
        </button>
      </div>
    </div>

    <!-- 近 7 日流量 -->
    <div class="card">
      <div class="card-head">
        <h3>近 7 日流量统计</h3>
        <div class="legend">
          <span class="legend-item"><i class="legend-dot up"></i>上传</span>
          <span class="legend-item"><i class="legend-dot down"></i>下载</span>
        </div>
      </div>
      <div v-if="!store.loggedIn" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 12h4l3-8 4 16 3-8h4" />
        </svg>
        <p>登录后查看近 7 日上传下载流量</p>
        <button class="btn btn-ghost btn-sm" @click="$emit('open-login')">登录查看</button>
      </div>
      <div v-else-if="loadingTraffic" class="empty-state">
        <p>流量数据加载中...</p>
      </div>
      <div v-else class="traffic-chart">
        <div class="chart-bars">
          <div v-for="item in traffic" :key="item.date" class="bar-group" :title="trafficTooltip(item)">
            <div class="bar-track">
              <div class="bar up" :style="{ height: barHeight(item.out_bytes) }"></div>
              <div class="bar down" :style="{ height: barHeight(item.in_bytes) }"></div>
            </div>
            <span class="bar-label">{{ shortDate(item.date) }}</span>
          </div>
        </div>
        <div class="chart-totals">
          <span>近 7 日上传 <strong class="mono">{{ formatBytes(totalOut) }}</strong></span>
          <span>近 7 日下载 <strong class="mono">{{ formatBytes(totalIn) }}</strong></span>
        </div>
      </div>
    </div>

    <div class="home-grid">
      <!-- 公告 -->
      <div class="card">
        <div class="card-head">
          <h3>公告</h3>
        </div>
        <div v-if="announcements.length === 0" class="empty-state" style="padding: 24px;">
          <p>暂无公告</p>
        </div>
        <ul v-else class="announce-list">
          <li v-for="a in announcements" :key="a.id" class="announce-item">
            <span class="announce-title">{{ a.title }}</span>
            <span class="announce-date">{{ shortDate(a.created_at) }}</span>
          </li>
        </ul>
      </div>

      <!-- 功能入口 -->
      <div class="card">
        <div class="card-head">
          <h3>功能入口</h3>
        </div>
        <div class="quick-grid">
          <button class="quick-item" @click="go('/tunnels')">
            <span class="quick-icon blue">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7h16" /><path d="M4 12h10" /><path d="M4 17h16" /></svg>
            </span>
            <span>隧道管理</span>
          </button>
          <button class="quick-item" @click="go('/logs')">
            <span class="quick-icon green">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16v16H4z" /><path d="M8 9h8" /><path d="M8 13h8" /><path d="M8 17h5" /></svg>
            </span>
            <span>运行日志</span>
          </button>
          <button class="quick-item" @click="go('/settings')">
            <span class="quick-icon purple">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1.03 1.56V21a2 2 0 1 1-4 0v-.09A1.7 1.7 0 0 0 9 19.36a1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.7 1.7 0 0 0 .34-1.87 1.7 1.7 0 0 0-1.56-1.03H3a2 2 0 1 1 0-4h.09A1.7 1.7 0 0 0 4.64 9a1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.7 1.7 0 0 0 1.87.34h.09a1.7 1.7 0 0 0 1.03-1.56V3a2 2 0 1 1 4 0v.09A1.7 1.7 0 0 0 15 4.64a1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.7 1.7 0 0 0-.34 1.87v.09a1.7 1.7 0 0 0 1.56 1.03H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.51.98Z" /></svg>
            </span>
            <span>应用设置</span>
          </button>
          <button class="quick-item" @click="go('/tunnels')">
            <span class="quick-icon orange">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14" /><path d="M5 12h14" /></svg>
            </span>
            <span>新建隧道</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 帮助与反馈 -->
    <div class="card">
      <div class="card-head">
        <h3>帮助与反馈</h3>
      </div>
      <div class="help-grid">
        <a class="help-item" href="https://weavenet.cqfishnet.cn/docs" target="_blank" rel="noreferrer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 8v.01" /><path d="M12 12v5" /></svg>
          <div><strong>常见问题</strong><small>使用文档与 FAQ</small></div>
        </a>
        <a class="help-item" href="https://weavenet.cqfishnet.cn/help" target="_blank" rel="noreferrer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><path d="M7 10l5 5 5-5" /><path d="M12 15V3" /></svg>
          <div><strong>意见征集</strong><small>反馈问题或提出建议</small></div>
        </a>
        <a class="help-item" href="https://weavenet.cqfishnet.cn/announcements" target="_blank" rel="noreferrer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><path d="M3 11l18-8v18L3 13v-2z" /><path d="M3 13l4 2 14-12" /></svg>
          <div><strong>意见反馈</strong><small>提交工单获得支持</small></div>
        </a>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { api } from '../api';
import { store, formatBytes } from '../store';

const emit = defineEmits(['open-login']);
const router = useRouter();

const traffic = ref([]);
const loadingTraffic = ref(false);
const quota = ref({});
const announcements = ref([]);
const signing = ref(false);
const signedToday = ref(false);

const totalIn = computed(() => traffic.value.reduce((s, i) => s + (i.in_bytes || 0), 0));
const totalOut = computed(() => traffic.value.reduce((s, i) => s + (i.out_bytes || 0), 0));

function shortDate(str) {
  if (!str) return '';
  return String(str).slice(5).replace('-', '/');
}

function trafficTooltip(item) {
  return `${item.date}\n上传 ${formatBytes(item.out_bytes)}\n下载 ${formatBytes(item.in_bytes)}`;
}

function barHeight(bytes) {
  const max = Math.max(totalIn.value, totalOut.value, 1);
  const pct = (Number(bytes) || 0) / max;
  return `${Math.max(2, Math.round(pct * 100))}%`;
}

function go(path) {
  router.push(path);
}

async function loadTraffic() {
  if (!store.loggedIn) return;
  loadingTraffic.value = true;
  try {
    const data = await api.getTraffic(7);
    traffic.value = Array.isArray(data) ? data : [];
  } catch (err) {
    // 忽略
  } finally {
    loadingTraffic.value = false;
  }
}

async function loadQuota() {
  if (!store.loggedIn) return;
  try {
    quota.value = (await api.getQuota()) || {};
  } catch (err) {
    // 忽略
  }
}

async function loadAnnouncements() {
  try {
    const data = await api.getAnnouncements();
    announcements.value = (data && data.items) || [];
  } catch (err) {
    // 忽略
  }
}

async function loadSigninStatus() {
  if (!store.loggedIn) return;
  try {
    const st = await api.getSigninStatus();
    signedToday.value = !!(st && st.can_signin === false);
  } catch (err) {
    // 忽略
  }
}

async function doSignin() {
  if (!store.loggedIn || signing.value || signedToday.value) return;
  signing.value = true;
  try {
    const data = await api.signin();
    signedToday.value = true;
    if (store.user) store.user.points = (store.user.points || 0) + (data && data.gained_points || 0);
    // 简单提示
  } catch (err) {
    // 忽略
  } finally {
    signing.value = false;
  }
}

onMounted(() => {
  loadAnnouncements();
  loadTraffic();
  loadQuota();
  loadSigninStatus();
});
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.home-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.12), rgba(6, 182, 212, 0.08));
}

.hero-left h2 {
  font-size: 19px;
  margin-bottom: 6px;
}

.hero-desc {
  color: var(--text-sub);
  font-size: 13px;
}

.hero-stats {
  display: flex;
  gap: 26px;
  margin-top: 18px;
}

.hero-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-num {
  font-size: 21px;
  font-weight: 800;
  background: var(--gradient);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.stat-label {
  font-size: 12px;
  color: var(--text-faint);
}

.home-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}

.legend {
  display: flex;
  gap: 14px;
  font-size: 12px;
  color: var(--text-sub);
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.legend-dot {
  width: 9px;
  height: 9px;
  border-radius: 3px;
}

.legend-dot.up {
  background: var(--accent);
}

.legend-dot.down {
  background: var(--primary);
}

.traffic-chart {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  gap: 14px;
  height: 140px;
  padding-top: 8px;
}

.bar-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
  height: 100%;
}

.bar-track {
  flex: 1;
  width: 100%;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 3px;
  background: rgba(138, 164, 187, 0.08);
  border-radius: 6px;
  overflow: hidden;
}

.bar {
  width: 45%;
  border-radius: 3px 3px 0 0;
  min-height: 2px;
  transition: height 0.4s;
}

.bar.up {
  background: linear-gradient(180deg, var(--accent), rgba(6, 182, 212, 0.4));
}

.bar.down {
  background: linear-gradient(180deg, var(--primary), rgba(14, 165, 233, 0.35));
}

.bar-label {
  font-size: 11px;
  color: var(--text-faint);
}

.chart-totals {
  display: flex;
  gap: 24px;
  font-size: 12.5px;
  color: var(--text-sub);
}

.chart-totals strong {
  color: var(--text);
}

.announce-list {
  list-style: none;
  display: flex;
  flex-direction: column;
}

.announce-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 2px;
  border-bottom: 1px dashed var(--border);
}

.announce-item:last-child {
  border-bottom: none;
}

.announce-title {
  font-size: 13.5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.announce-date {
  font-size: 12px;
  color: var(--text-faint);
  flex-shrink: 0;
  margin-left: 12px;
}

.quick-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.quick-item {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--surface);
  cursor: pointer;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text);
  font-family: inherit;
  transition: border-color 0.15s, background 0.15s;
  text-align: left;
}

.quick-item:hover {
  border-color: var(--primary);
  background: var(--surface-strong);
}

.quick-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.quick-icon svg {
  width: 17px;
  height: 17px;
}

.quick-icon.blue {
  background: rgba(14, 165, 233, 0.14);
  color: var(--primary);
}

.quick-icon.green {
  background: rgba(16, 185, 129, 0.14);
  color: var(--success);
}

.quick-icon.purple {
  background: rgba(139, 92, 246, 0.14);
  color: #8b5cf6;
}

.quick-icon.orange {
  background: rgba(245, 158, 11, 0.14);
  color: var(--warning);
}

.help-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.help-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--surface);
  text-decoration: none;
  color: var(--text);
  transition: border-color 0.15s;
}

.help-item:hover {
  border-color: var(--primary);
}

.help-item svg {
  width: 22px;
  height: 22px;
  color: var(--primary);
  flex-shrink: 0;
}

.help-item div {
  display: flex;
  flex-direction: column;
  line-height: 1.4;
}

.help-item strong {
  font-size: 13.5px;
}

.help-item small {
  font-size: 12px;
  color: var(--text-faint);
}

@media (max-width: 900px) {
  .home-grid {
    grid-template-columns: 1fr;
  }

  .help-grid {
    grid-template-columns: 1fr;
  }
}
</style>
