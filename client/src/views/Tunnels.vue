<template>
  <div class="tunnels">
    <!-- 未登录提示 -->
    <div v-if="!store.loggedIn" class="login-tip card">
      <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21c1.5-3.5 4.5-5 8-5s6.5 1.5 8 5" />
      </svg>
      <span>登录后可查看并管理你的隧道</span>
      <button class="btn btn-primary btn-sm" @click="$emit('open-login')">立即登录</button>
    </div>

    <div class="tunnels-head">
      <div>
        <h2>我的隧道</h2>
        <p class="head-sub">共 {{ store.tunnels.length }} 条 · 支持 TCP / UDP / HTTP / HTTPS 等协议</p>
      </div>
      <button class="btn btn-primary" :disabled="!store.loggedIn" @click="openCreate">
        <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14" /><path d="M5 12h14" /></svg>
        新建隧道
      </button>
    </div>

    <!-- 空态引导 -->
    <div v-if="store.loggedIn && store.tunnels.length === 0" class="card empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 7h16" />
        <path d="M4 12h10" />
        <path d="M4 17h16" />
      </svg>
      <p>还没有隧道，点击右上角「新建隧道」创建第一条吧</p>
      <button class="btn btn-primary" @click="openCreate">新建隧道</button>
    </div>

    <!-- 隧道列表 -->
    <div v-else-if="store.tunnels.length > 0" class="tunnel-list">
      <div
        v-for="t in store.tunnels"
        :key="t.id"
        class="tunnel-card glass"
        :class="{ active: selectedId === t.id }"
        @click="select(t.id)"
      >
        <div class="tunnel-card-main">
          <div class="tunnel-card-title">
            <span class="type-tag">{{ t.type }}</span>
            <strong>{{ t.name }}</strong>
            <span class="status-badge" :class="statusInfo(t).cls">{{ statusInfo(t).text }}</span>
          </div>
          <div class="tunnel-card-addr mono">{{ t.public_address || '-' }}</div>
          <div class="tunnel-card-sub">
            <span>节点 {{ t.node_name || '-' }}</span>
            <span>本地 {{ t.local_ip || '127.0.0.1' }}:{{ t.local_port ?? '-' }}</span>
            <span>今日 {{ formatBytes(runtimeOf(t.id).today_in || 0) }}↓ / {{ formatBytes(runtimeOf(t.id).today_out || 0) }}↑</span>
          </div>
        </div>
        <div class="tunnel-card-ops" @click.stop>
          <button class="btn btn-sm" :class="statusInfo(t).text === '运行中' ? 'btn-danger' : 'btn-ghost'"
                  :disabled="!store.loggedIn || busy" @click="toggle(t)">
            {{ statusInfo(t).text === '运行中' ? '停止' : '启动' }}
          </button>
          <button class="btn btn-sm btn-ghost" :disabled="!store.loggedIn" @click="openDetail(t)">详情</button>
        </div>
      </div>
    </div>

    <!-- 未登录且无数据时也显示空态引导 -->
    <div v-else class="card empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 7h16" />
        <path d="M4 12h10" />
        <path d="M4 17h16" />
      </svg>
      <p>登录后查看和管理你的隧道</p>
      <button class="btn btn-primary" @click="$emit('open-login')">立即登录</button>
    </div>

    <!-- 隧道详情弹窗 -->
    <div v-if="detailTunnel" class="modal-mask" @click.self="detailTunnel = null">
      <div class="modal-panel glass" style="width: 520px;">
        <div class="modal-head">
          <h2>{{ detailTunnel.name }}</h2>
          <button class="icon-btn" @click="detailTunnel = null">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
          </button>
        </div>
        <div class="detail-grid">
          <div class="detail-item"><span class="dl">类型</span><span class="dv type-tag">{{ detailTunnel.type }}</span></div>
          <div class="detail-item"><span class="dl">状态</span><span class="status-badge" :class="statusInfo(detailTunnel).cls">{{ statusInfo(detailTunnel).text }}</span></div>
          <div class="detail-item"><span class="dl">本地地址</span><span class="dv mono">{{ detailTunnel.local_ip || '127.0.0.1' }}:{{ detailTunnel.local_port ?? '-' }}</span></div>
          <div class="detail-item"><span class="dl">远程端口</span><span class="dv mono">{{ detailTunnel.remote_port ?? (detailTunnel.subdomain || detailTunnel.custom_domain || '-') }}</span></div>
          <div class="detail-item"><span class="dl">公网地址</span><span class="dv mono">{{ detailTunnel.public_address || '-' }}</span></div>
          <div class="detail-item"><span class="dl">所属节点</span><span class="dv">{{ detailTunnel.node_name || '-' }}</span></div>
          <div class="detail-item"><span class="dl">限速</span><span class="dv">{{ (detailTunnel.bandwidth_limit_kbps || 0) / 1000 }} Mbps</span></div>
          <div class="detail-item"><span class="dl">加密/压缩/KCP</span><span class="dv">{{ flags(detailTunnel) }}</span></div>
          <div class="detail-item"><span class="dl">今日上行</span><span class="dv mono">{{ formatBytes(runtimeOf(detailTunnel.id).today_out || 0) }}</span></div>
          <div class="detail-item"><span class="dl">今日下行</span><span class="dv mono">{{ formatBytes(runtimeOf(detailTunnel.id).today_in || 0) }}</span></div>
        </div>
        <div class="detail-config">
          <div class="config-head">
            <h3>配置预览</h3>
            <button class="text-btn" @click="loadConfig">刷新</button>
          </div>
          <pre class="config-preview">{{ configText || '加载中...' }}</pre>
        </div>
        <div class="modal-foot">
          <button class="btn btn-danger" :disabled="busy" @click="removeTunnel(detailTunnel)">删除隧道</button>
          <button class="btn btn-ghost" @click="detailTunnel = null">关闭</button>
        </div>
      </div>
    </div>

    <!-- 新建隧道弹窗 -->
    <div v-if="showCreate" class="modal-mask" @click.self="showCreate = false">
      <div class="modal-panel glass" style="width: 560px; max-height: 86vh; overflow-y: auto;">
        <div class="modal-head">
          <h2>新建隧道</h2>
          <button class="icon-btn" @click="showCreate = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
          </button>
        </div>

        <div class="field">
          <label>隧道名称</label>
          <input v-model="form.name" type="text" spellcheck="false" placeholder="例如：我的网站" />
        </div>
        <div class="form-row">
          <div class="field">
            <label>协议类型</label>
            <select v-model="form.type">
              <option value="tcp">TCP</option>
              <option value="udp">UDP</option>
              <option value="http">HTTP</option>
              <option value="https">HTTPS</option>
              <option value="stcp">STCP</option>
              <option value="xtcp">XTCP</option>
              <option value="kcp">KCP</option>
            </select>
          </div>
          <div class="field">
            <label>节点</label>
            <select v-model="form.node_id">
              <option v-for="n in nodes" :key="n.id" :value="n.id" :disabled="n.status !== 'online'">
                {{ n.name }}{{ n.status !== 'online' ? '（离线）' : '' }}
              </option>
            </select>
          </div>
        </div>
        <div class="form-row">
          <div class="field">
            <label>本地 IP</label>
            <input v-model="form.local_ip" type="text" spellcheck="false" placeholder="127.0.0.1" />
          </div>
          <div class="field">
            <label>本地端口</label>
            <input v-model.number="form.local_port" type="number" min="1" max="65535" placeholder="例如 8080" />
          </div>
        </div>
        <div class="form-row">
          <div class="field" v-if="isTcpLike">
            <label>远程端口（留空自动分配）</label>
            <input v-model.number="form.remote_port" type="number" min="1" max="65535" placeholder="自动" />
          </div>
          <div class="field" v-if="isHttpLike">
            <label>子域名（可选）</label>
            <input v-model="form.subdomain" type="text" spellcheck="false" placeholder="自定义前缀" />
          </div>
          <div class="field" v-if="isHttpLike">
            <label>自定义域名（可选）</label>
            <input v-model="form.custom_domain" type="text" spellcheck="false" placeholder="yourdomain.com" />
          </div>
          <div class="field" v-if="isStcpLike">
            <label>密钥（留空自动生成）</label>
            <input v-model="form.secret_key" type="text" spellcheck="false" placeholder="自动" />
          </div>
        </div>
        <div class="option-row">
          <label class="check-item">
            <input type="checkbox" v-model="form.encryption" /> 加密
          </label>
          <label class="check-item">
            <input type="checkbox" v-model="form.compression" /> 压缩
          </label>
          <label class="check-item">
            <input type="checkbox" v-model="form.kcp" /> KCP
          </label>
        </div>
        <p v-if="createError" class="form-error">{{ createError }}</p>
        <div class="modal-foot">
          <button class="btn btn-ghost" @click="showCreate = false">取消</button>
          <button class="btn btn-primary" :disabled="creating" @click="doCreate">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { api } from '../api';
import { store, statusInfo, runtimeOf, formatBytes } from '../store';

const emit = defineEmits(['open-login']);

const selectedId = ref(null);
const busy = ref(false);
const detailTunnel = ref(null);
const configText = ref('');
const showCreate = ref(false);
const creating = ref(false);
const createError = ref('');
const nodes = ref([]);

const emptyForm = () => ({
  name: '',
  node_id: null,
  type: 'tcp',
  local_ip: '127.0.0.1',
  local_port: 8080,
  remote_port: null,
  subdomain: '',
  custom_domain: '',
  secret_key: '',
  kcp: false,
  encryption: true,
  compression: false,
});
const form = ref(emptyForm());

const isTcpLike = computed(() => ['tcp', 'udp', 'kcp'].includes(form.value.type));
const isHttpLike = computed(() => ['http', 'https'].includes(form.value.type));
const isStcpLike = computed(() => ['stcp', 'xtcp'].includes(form.value.type));

function flags(t) {
  const f = [];
  if (t.encryption) f.push('加密');
  if (t.compression) f.push('压缩');
  if (t.kcp) f.push('KCP');
  return f.length ? f.join(' / ') : '无';
}

function select(id) {
  selectedId.value = id;
}

async function toggle(t) {
  if (!store.loggedIn || busy.value) return;
  busy.value = true;
  try {
    if (statusInfo(t).text === '运行中') {
      await api.stopTunnel(t.id);
    } else {
      await api.startTunnel(t.id);
    }
  } catch (err) {
    // 主进程 toast 处理
  } finally {
    busy.value = false;
  }
}

async function openDetail(t) {
  detailTunnel.value = t;
  configText.value = '加载中...';
  await loadConfig();
}

async function loadConfig() {
  if (!detailTunnel.value) return;
  try {
    configText.value = (await api.getTunnelConfig(detailTunnel.value.id)) || '（面板未返回配置内容）';
  } catch (err) {
    configText.value = `配置拉取失败：${err.message}`;
  }
}

async function removeTunnel(t) {
  if (!store.loggedIn || busy.value) return;
  if (!confirm(`确定删除隧道「${t.name}」？`)) return;
  busy.value = true;
  try {
    await api.deleteTunnel(t.id);
    store.tunnels = store.tunnels.filter((x) => x.id !== t.id);
    if (detailTunnel.value && detailTunnel.value.id === t.id) detailTunnel.value = null;
  } catch (err) {
    alert(err.message);
  } finally {
    busy.value = false;
  }
}

async function openCreate() {
  if (!store.loggedIn) {
    emit('open-login');
    return;
  }
  form.value = emptyForm();
  createError.value = '';
  showCreate.value = true;
  try {
    nodes.value = (await api.getNodes()) || [];
    if (nodes.value.length > 0) {
      const online = nodes.value.find((n) => n.status === 'online');
      form.value.node_id = online ? online.id : nodes.value[0].id;
    }
  } catch (err) {
    createError.value = err.message;
  }
}

async function doCreate() {
  if (!form.value.name) {
    createError.value = '请填写隧道名称';
    return;
  }
  if (!form.value.node_id) {
    createError.value = nodes.value.length === 0 ? '暂无可用节点，请联系管理员配置' : '请选择在线节点';
    return;
  }
  if (!form.value.local_port) {
    createError.value = '请填写本地端口';
    return;
  }
  creating.value = true;
  createError.value = '';
  const payload = {
    name: form.value.name,
    node_id: form.value.node_id,
    type: form.value.type,
    local_ip: form.value.local_ip || '127.0.0.1',
    local_port: form.value.local_port,
    kcp: form.value.kcp,
    encryption: form.value.encryption,
    compression: form.value.compression,
  };
  if (isTcpLike.value && form.value.remote_port) payload.remote_port = form.value.remote_port;
  if (isHttpLike.value) {
    if (form.value.subdomain) payload.subdomain = form.value.subdomain;
    if (form.value.custom_domain) payload.custom_domain = form.value.custom_domain;
  }
  if (isStcpLike.value && form.value.secret_key) payload.secret_key = form.value.secret_key;
  try {
    await api.createTunnel(payload);
    showCreate.value = false;
    const data = await api.getTunnels();
    store.tunnels = (data && data.tunnels) || [];
  } catch (err) {
    createError.value = err.message;
  } finally {
    creating.value = false;
  }
}

onMounted(() => {
  if (store.loggedIn) {
    api.getNodes().then((d) => (nodes.value = d || [])).catch(() => {});
  }
});
</script>

<style scoped>
.tunnels {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.tunnels-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.tunnels-head h2 {
  font-size: 17px;
}

.head-sub {
  font-size: 12.5px;
  color: var(--text-faint);
  margin-top: 3px;
}

.tunnel-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tunnel-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.tunnel-card:hover,
.tunnel-card.active {
  border-color: var(--primary);
  background: var(--surface-strong);
}

.tunnel-card-main {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tunnel-card-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tunnel-card-title strong {
  font-size: 14.5px;
}

.tunnel-card-addr {
  font-size: 13px;
  color: var(--primary-deep);
}

.theme-dark .tunnel-card-addr {
  color: #7dd3fc;
}

.tunnel-card-sub {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-faint);
  flex-wrap: wrap;
}

.tunnel-card-ops {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.option-row {
  display: flex;
  gap: 18px;
  margin-bottom: 14px;
}

.check-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-sub);
  cursor: pointer;
}

.check-item input {
  accent-color: var(--primary);
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 18px;
  margin-bottom: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.detail-item .dl {
  font-size: 11.5px;
  color: var(--text-faint);
}

.detail-item .dv {
  font-size: 13px;
  word-break: break-all;
}

.detail-config {
  margin-bottom: 16px;
}

.config-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.config-head h3 {
  font-size: 13.5px;
}

.text-btn {
  background: none;
  border: none;
  color: var(--primary);
  font-size: 12.5px;
  cursor: pointer;
  font-family: inherit;
}

.config-preview {
  max-height: 220px;
  overflow: auto;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px;
  font-family: 'Cascadia Mono', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-all;
}

.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
}
</style>
