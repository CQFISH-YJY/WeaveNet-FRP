<template>
  <div class="logs">
    <!-- 未登录提示 -->
    <div v-if="!store.loggedIn" class="login-tip card">
      <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21c1.5-3.5 4.5-5 8-5s6.5 1.5 8 5" />
      </svg>
      <span>登录后可查看隧道运行日志</span>
      <button class="btn btn-primary btn-sm" @click="$emit('open-login')">立即登录</button>
    </div>

    <div class="logs-head">
      <div>
        <h2>运行日志</h2>
        <p class="head-sub">查看各隧道 frpc 进程运行日志</p>
      </div>
      <div class="logs-ops">
        <select v-model="filterTunnelId" class="filter-select">
          <option value="">全部隧道</option>
          <option v-for="t in store.tunnels" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
        <label class="check-item">
          <input type="checkbox" v-model="autoScroll" /> 自动滚动
        </label>
        <button class="btn btn-ghost btn-sm" @click="clearLogs">清空</button>
        <button class="btn btn-ghost btn-sm" @click="saveLogs">保存到本地</button>
      </div>
    </div>

    <div class="card log-panel">
      <div v-if="!store.loggedIn" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 4h16v16H4z" />
          <path d="M8 9h8" />
          <path d="M8 13h8" />
          <path d="M8 17h5" />
        </svg>
        <p>登录后查看日志</p>
      </div>
      <div v-else-if="filteredLogs.length === 0" class="empty-state">
        <p>暂无日志，启动隧道后会自动记录运行日志</p>
      </div>
      <div v-else class="log-lines" ref="logBox">
        <div v-for="(line, i) in filteredLogs" :key="i" class="log-line" :class="lineClass(line)">
          <span class="log-time">{{ line.time }}</span>
          <span class="log-level">{{ line.level }}</span>
          <span class="log-tag" v-if="line.tunnelName">{{ line.tunnelName }}</span>
          <span class="log-msg">{{ line.msg }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue';
import { api } from '../api';
import { store } from '../store';

const emit = defineEmits(['open-login']);

const filterTunnelId = ref('');
const autoScroll = ref(true);
const logs = ref([]);
const logBox = ref(null);

let pollTimer = null;

const filteredLogs = computed(() => {
  if (!filterTunnelId.value) return logs.value;
  return logs.value.filter((l) => String(l.tunnelId) === String(filterTunnelId.value));
});

function lineClass(line) {
  if (!line.level) return '';
  const level = String(line.level).toLowerCase();
  if (['error', 'fatal', 'panic'].includes(level)) return 'level-error';
  if (['warn', 'warning'].includes(level)) return 'level-warn';
  if (['debug'].includes(level)) return 'level-debug';
  return '';
}

function clearLogs() {
  logs.value = [];
}

function saveLogs() {
  const content = filteredLogs.value
    .map((l) => `[${l.time}] [${l.level || 'info'}]${l.tunnelName ? ` [${l.tunnelName}]` : ''} ${l.msg}`)
    .join('\n');
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = `weavenet-logs-${Date.now()}.txt`;
  a.click();
  URL.revokeObjectURL(a.href);
}

async function pollLogs() {
  if (!store.loggedIn) return;
  try {
    const data = await api.getTunnelLogs(null, 200);
    if (Array.isArray(data)) {
      logs.value = data.map((item) => ({
        time: item.time || '',
        level: item.level || 'info',
        msg: item.msg || '',
        tunnelId: item.tunnelId,
        tunnelName: item.tunnelName || '',
      }));
    }
  } catch (err) {
    // 忽略
  }
}

watch(filteredLogs, async () => {
  if (autoScroll.value) {
    await nextTick();
    if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
  }
});

onMounted(() => {
  pollLogs();
  if (store.loggedIn) {
    pollTimer = setInterval(pollLogs, 2000);
  }
});

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});
</script>

<style scoped>
.logs {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.logs-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.logs-head h2 {
  font-size: 17px;
}

.head-sub {
  font-size: 12.5px;
  color: var(--text-faint);
  margin-top: 3px;
}

.logs-ops {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-select {
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-strong);
  background: var(--surface-strong);
  color: var(--text);
  font-size: 12.5px;
  font-family: inherit;
  outline: none;
}

.check-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12.5px;
  color: var(--text-sub);
  cursor: pointer;
}

.check-item input {
  accent-color: var(--primary);
}

.log-panel {
  height: calc(100vh - 220px);
  min-height: 320px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.log-lines {
  flex: 1;
  overflow-y: auto;
  padding: 14px 16px;
  font-family: 'Cascadia Mono', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.7;
}

.log-line {
  display: flex;
  gap: 10px;
  white-space: pre-wrap;
  word-break: break-all;
  padding: 1px 0;
  color: var(--text-sub);
}

.log-time {
  color: var(--text-faint);
  flex-shrink: 0;
}

.log-level {
  flex-shrink: 0;
  width: 38px;
  text-align: center;
  border-radius: 4px;
  font-weight: 700;
  font-size: 11px;
  line-height: 1.6;
  background: rgba(138, 164, 187, 0.14);
  color: var(--text-faint);
}

.log-tag {
  flex-shrink: 0;
  padding: 0 7px;
  border-radius: 4px;
  background: rgba(14, 165, 233, 0.14);
  color: var(--primary);
  font-weight: 600;
  font-size: 11px;
  line-height: 1.6;
}

.log-msg {
  color: var(--text);
}

.level-error .log-level {
  background: rgba(239, 68, 68, 0.16);
  color: #dc2626;
}

.level-error .log-msg {
  color: #dc2626;
}

.level-warn .log-level {
  background: rgba(245, 158, 11, 0.16);
  color: #b45309;
}

.level-warn .log-msg {
  color: #b45309;
}

.level-debug .log-level {
  background: rgba(138, 164, 187, 0.14);
  color: var(--text-faint);
}

.level-debug .log-msg {
  color: var(--text-faint);
}
</style>
