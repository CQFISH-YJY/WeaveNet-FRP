<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>隧道管理</h2>
        <div class="sub">创建并管理你的内网穿透隧道</div>
      </div>
      <n-button class="btn-grad-orange" @click="openCreate">
        <template #icon>
          <svg-icon name="plus" :size="16" />
        </template>
        创建隧道
      </n-button>
    </div>

    <div class="glass-card table-wrap">
      <n-table :bordered="false" :single-line="false" size="small">
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>节点</th>
            <th>本地地址</th>
            <th>远程端口</th>
            <th>公网地址</th>
            <th>状态</th>
            <th>流量</th>
            <th style="width: 260px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in tunnels" :key="t.id">
            <td class="cell-main">
              <div class="tunnel-name">{{ t.name }}</div>
              <div class="cell-sub">{{ t.id }} · {{ t.local_ip || '-' }}</div>
            </td>
            <td><span class="type-tag">{{ tunnelTypeText(t.type) }}</span></td>
            <td>{{ t.node_name || t.node?.name || '-' }}</td>
            <td>{{ t.local_ip || '-' }}:{{ t.local_port || '-' }}</td>
            <td>{{ t.remote_port || '自动' }}</td>
            <td class="cell-main">
              <div class="public-url">{{ publicUrl(t) }}</div>
              <div class="cell-sub">{{ t.subdomain || t.custom_domain || '' }}</div>
            </td>
            <td>
              <status-badge :type="tunnelStatusType(t.online ?? t.status)" :text="tunnelStatusText(t.online ?? t.status)" />
            </td>
            <td>
              <div class="cell-sub">入 {{ formatBytes(t.traffic_in) }}</div>
              <div class="cell-sub">出 {{ formatBytes(t.traffic_out) }}</div>
            </td>
            <td>
              <div class="table-actions">
                <n-button v-if="!isRunning(t)" size="tiny" type="success" ghost @click="handleStart(t)">
                  <template #icon><svg-icon name="play" :size="13" /></template>
                  启动
                </n-button>
                <n-button v-else size="tiny" type="warning" ghost @click="handleStop(t)">
                  <template #icon><svg-icon name="stop" :size="13" /></template>
                  停止
                </n-button>
                <n-button size="tiny" @click="openConfig(t)">
                  <template #icon><svg-icon name="copy" :size="13" /></template>
                  配置
                </n-button>
                <n-button size="tiny" @click="openEdit(t)">
                  <template #icon><svg-icon name="edit" :size="13" /></template>
                  编辑
                </n-button>
                <n-button size="tiny" type="error" ghost @click="handleDelete(t)">
                  <template #icon><svg-icon name="trash" :size="13" /></template>
                  删除
                </n-button>
              </div>
            </td>
          </tr>
          <tr v-if="!tunnels.length">
            <td colspan="9">
              <div class="empty-tip">
                <div style="margin-bottom: 10px">还没有创建任何隧道</div>
                <n-button class="btn-grad-cyan" size="small" @click="openCreate">立即创建第一条隧道</n-button>
              </div>
            </td>
          </tr>
        </tbody>
      </n-table>
    </div>

    <n-modal
      v-model:show="modalVisible"
      :mask-closable="false"
      preset="card"
      :title="editing ? '编辑隧道' : '创建隧道'"
      style="width: 560px; max-width: 94vw"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top" size="medium">
        <div class="form-grid">
          <n-form-item label="隧道名称" path="name" style="grid-column: span 2">
            <n-input v-model:value="form.name" placeholder="例如：我的网站" clearable />
          </n-form-item>
          <n-form-item label="隧道类型" path="type">
            <n-select v-model:value="form.type" :options="typeOptions" />
          </n-form-item>
          <n-form-item label="穿透节点" path="node_id">
            <n-select
              v-model:value="form.node_id"
              :options="nodeOptions"
              :loading="nodesLoading"
              placeholder="选择节点"
            />
          </n-form-item>
          <n-form-item label="本地 IP" path="local_ip">
            <n-input v-model:value="form.local_ip" placeholder="127.0.0.1" />
          </n-form-item>
          <n-form-item label="本地端口" path="local_port">
            <n-input-number v-model:value="form.local_port" :min="1" :max="65535" style="width: 100%" placeholder="例如 8080" />
          </n-form-item>
          <n-form-item label="远程端口" path="remote_port">
            <n-input-number
              v-model:value="form.remote_port"
              :min="1"
              :max="65535"
              style="width: 100%"
              placeholder="留空自动分配"
            />
          </n-form-item>
          <n-form-item v-if="isHttpLike(form.type)" label="子域名（http/https）" path="subdomain">
            <n-input v-model:value="form.subdomain" placeholder="例如 myweb" />
          </n-form-item>
          <n-form-item v-if="isHttpLike(form.type)" label="自定义域名（可选）" path="custom_domain">
            <n-input v-model:value="form.custom_domain" placeholder="例如 www.example.com" />
          </n-form-item>
        </div>
        <div class="switch-row">
          <n-checkbox v-model:checked="form.kcp">启用 KCP 加速</n-checkbox>
          <n-checkbox v-model:checked="form.encryption">加密传输</n-checkbox>
          <n-checkbox v-model:checked="form.compression">压缩传输</n-checkbox>
        </div>
        <div class="modal-actions">
          <n-button @click="modalVisible = false">取消</n-button>
          <n-button class="btn-grad-cyan" :loading="saving" @click="handleSave">保存</n-button>
        </div>
      </n-form>
    </n-modal>

    <n-modal v-model:show="configVisible" preset="card" title="隧道配置" style="width: 620px; max-width: 94vw">
      <div class="code-block">{{ configText }}</div>
      <div class="modal-actions">
        <n-button class="btn-grad-cyan" @click="copyConfig">复制配置</n-button>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import {
  useMessage,
  useDialog,
  NButton,
  NTable,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NCheckbox
} from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import {
  getTunnels,
  getNodes,
  createTunnel,
  updateTunnel,
  deleteTunnel,
  startTunnel,
  stopTunnel,
  getTunnelConfig
} from '@/api'
import { formatBytes, tunnelTypeText, tunnelStatusText, tunnelStatusType } from '@/utils/format'

const message = useMessage()
const dialog = useDialog()

const tunnels = ref([])
const nodes = ref([])
const nodesLoading = ref(false)
const modalVisible = ref(false)
const configVisible = ref(false)
const editing = ref(null)
const saving = ref(false)
const configText = ref('')
const formRef = ref(null)

const emptyForm = () => ({
  name: '',
  type: 'tcp',
  node_id: null,
  local_ip: '127.0.0.1',
  local_port: null,
  remote_port: null,
  subdomain: '',
  custom_domain: '',
  kcp: false,
  encryption: false,
  compression: false
})

const form = reactive(emptyForm())

const typeOptions = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
  { label: 'STCP 安全隧道', value: 'stcp' },
  { label: 'XTCP 点对点', value: 'xtcp' },
  { label: 'KCP 加速', value: 'kcp' },
  { label: '负载均衡', value: 'loadbalance' }
]

const nodeOptions = computed(() =>
  nodes.value.map((n) => ({
    label: `${n.name}${n.status === 'online' || n.status === 1 ? '' : '（离线）'}`,
    value: n.id
  }))
)

const rules = {
  name: [{ required: true, message: '请输入隧道名称', trigger: ['input', 'blur'] }],
  type: [{ required: true, message: '请选择隧道类型', trigger: ['change'] }],
  node_id: [{ required: true, message: '请选择穿透节点', trigger: ['change'] }],
  local_ip: [{ required: true, message: '请输入本地 IP', trigger: ['input', 'blur'] }],
  local_port: [{ required: true, message: '请输入本地端口', trigger: ['change'] }]
}

function isHttpLike(type) {
  return type === 'http' || type === 'https'
}

function isRunning(t) {
  const s = t.online ?? t.status
  return s === 'running' || s === 'online' || s === true || s === 1 || s === '1'
}

function publicUrl(t) {
  if (t.public_addr) return t.public_addr
  if (t.public_url) return t.public_url
  if (t.full_domain) return t.full_domain
  if (t.subdomain) return t.subdomain
  if (t.custom_domain) return t.custom_domain
  return '-'
}

async function loadTunnels() {
  try {
    tunnels.value = (await getTunnels()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function loadNodes() {
  nodesLoading.value = true
  try {
    nodes.value = (await getNodes()) || []
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    nodesLoading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, emptyForm())
  modalVisible.value = true
}

function openEdit(t) {
  editing.value = t
  Object.assign(form, {
    name: t.name || '',
    type: t.type || 'tcp',
    node_id: t.node_id ?? null,
    local_ip: t.local_ip || '127.0.0.1',
    local_port: t.local_port ?? null,
    remote_port: t.remote_port ?? null,
    subdomain: t.subdomain || '',
    custom_domain: t.custom_domain || '',
    kcp: !!t.kcp,
    encryption: !!t.encryption,
    compression: !!t.compression
  })
  modalVisible.value = true
}

async function handleSave() {
  try {
    await formRef.value?.validate()
  } catch (e) {
    return
  }
  saving.value = true
  const payload = {
    name: form.name,
    type: form.type,
    node_id: form.node_id,
    local_ip: form.local_ip,
    local_port: form.local_port
  }
  if (form.remote_port) payload.remote_port = form.remote_port
  if (form.subdomain) payload.subdomain = form.subdomain
  if (form.custom_domain) payload.custom_domain = form.custom_domain
  if (form.kcp) payload.kcp = true
  if (form.encryption) payload.encryption = true
  if (form.compression) payload.compression = true
  try {
    if (editing.value) {
      await updateTunnel(editing.value.id, payload)
      message.success('隧道已更新')
    } else {
      await createTunnel(payload)
      message.success('隧道创建成功')
    }
    modalVisible.value = false
    await loadTunnels()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

async function handleStart(t) {
  try {
    await startTunnel(t.id)
    message.success('隧道已启动')
    await loadTunnels()
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function handleStop(t) {
  try {
    await stopTunnel(t.id)
    message.success('隧道已停止')
    await loadTunnels()
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function handleDelete(t) {
  dialog.warning({
    title: '删除隧道',
    content: `确定要删除隧道「${t.name}」吗？删除后不可恢复。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteTunnel(t.id)
        message.success('隧道已删除')
        await loadTunnels()
      } catch (e) {
        // 错误已由拦截器提示
      }
    }
  })
}

async function openConfig(t) {
  try {
    const data = await getTunnelConfig(t.id)
    configText.value = typeof data === 'string' ? data : data?.config || JSON.stringify(data, null, 2)
    configVisible.value = true
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function copyConfig() {
  try {
    await navigator.clipboard.writeText(configText.value)
    message.success('配置已复制到剪贴板')
  } catch (e) {
    message.error('复制失败，请手动选择复制')
  }
}

onMounted(() => {
  loadTunnels()
  loadNodes()
})
</script>

<style scoped>
.table-wrap {
  padding: 12px 16px;
}

.cell-main {
  min-width: 0;
}

.tunnel-name {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.cell-sub {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

.type-tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #0284c7;
  background: rgba(14, 165, 233, 0.12);
  border: 1px solid rgba(14, 165, 233, 0.2);
  white-space: nowrap;
}

.public-url {
  font-size: 12px;
  color: #0ea5e9;
  word-break: break-all;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 16px;
}

.switch-row {
  display: flex;
  align-items: center;
  gap: 20px;
  margin: 4px 0 18px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

.n-table .n-tbody-tr {
  vertical-align: middle;
}
</style>


