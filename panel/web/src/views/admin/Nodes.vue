<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>节点管理</h2>
        <div class="sub">管理 frps 穿透节点及其 Agent Token</div>
      </div>
      <n-button class="btn-grad-orange" @click="openCreate">
        <template #icon><svg-icon name="plus" :size="16" /></template>
        新增节点
      </n-button>
    </div>

    <div class="glass-card table-wrap">
      <n-table :bordered="false" size="small">
        <thead>
          <tr>
            <th>名称</th>
            <th>地址</th>
            <th>端口</th>
            <th>状态</th>
            <th>限速</th>
            <th>备注</th>
            <th>Agent Token</th>
            <th style="width: 300px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="n in nodes" :key="n.id">
            <td class="cell-main">
              <div class="node-name">{{ n.name }}</div>
            </td>
            <td>{{ n.address || '-' }}</td>
            <td>{{ n.port || '-' }}</td>
            <td>
              <status-badge :type="isOnline(n) ? 'success' : 'error'" :text="isOnline(n) ? '在线' : '离线'" />
            </td>
            <td>{{ n.speed_limit_mbps ?? '-' }} Mbps</td>
            <td class="cell-sub">{{ n.remark || '-' }}</td>
            <td>
              <div class="token-cell">
                <span class="token-text">{{ maskToken(n.agent_token) }}</span>
                <n-button v-if="n.agent_token" size="tiny" quaternary @click="copyToken(n)">
                  <template #icon><svg-icon name="copy" :size="13" /></template>
                </n-button>
              </div>
            </td>
            <td>
              <div class="table-actions">
                <n-button v-if="isOnline(n)" size="tiny" type="warning" ghost @click="handleStop(n)">停用</n-button>
                <n-button v-else size="tiny" type="success" ghost @click="handleStart(n)">启用</n-button>
                <n-button size="tiny" @click="openSpeed(n)">限速</n-button>
                <n-button size="tiny" @click="openEdit(n)">
                  <template #icon><svg-icon name="edit" :size="13" /></template>
                  编辑
                </n-button>
                <n-button size="tiny" type="error" ghost @click="handleDelete(n)">
                  <template #icon><svg-icon name="trash" :size="13" /></template>
                  删除
                </n-button>
              </div>
            </td>
          </tr>
          <tr v-if="!nodes.length">
            <td colspan="8">
              <div class="empty-tip">暂无节点，点击右上角新增</div>
            </td>
          </tr>
        </tbody>
      </n-table>
    </div>

    <n-modal v-model:show="modalVisible" preset="card" :title="editing ? '编辑节点' : '新增节点'" style="width: 520px; max-width: 94vw">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top">
        <n-form-item label="节点名称" path="name">
          <n-input v-model:value="form.name" placeholder="例如：上海一区" clearable />
        </n-form-item>
        <div class="form-grid">
          <n-form-item label="节点地址" path="address">
            <n-input v-model:value="form.address" placeholder="例如 frp.example.com" />
          </n-form-item>
          <n-form-item label="控制端口" path="port">
            <n-input-number v-model:value="form.port" :min="1" :max="65535" style="width: 100%" placeholder="7000" />
          </n-form-item>
          <n-form-item label="限速 (Mbps)" path="speed_limit_mbps">
            <n-input-number v-model:value="form.speed_limit_mbps" :min="0" style="width: 100%" placeholder="0 表示不限速" />
          </n-form-item>
          <n-form-item label="备注" path="remark">
            <n-input v-model:value="form.remark" placeholder="可选备注" />
          </n-form-item>
        </div>
      </n-form>
      <div class="modal-actions">
        <n-button @click="modalVisible = false">取消</n-button>
        <n-button class="btn-grad-cyan" :loading="saving" @click="handleSave">保存</n-button>
      </div>
    </n-modal>

    <n-modal v-model:show="speedVisible" preset="card" title="节点限速" style="width: 420px; max-width: 94vw">
      <div class="plan-user-tip">节点：{{ speedTarget?.name }}</div>
      <n-input-number v-model:value="speedForm.speed_limit_mbps" :min="0" style="width: 100%">
        <template #suffix>Mbps</template>
      </n-input-number>
      <div class="modal-actions">
        <n-button @click="speedVisible = false">取消</n-button>
        <n-button class="btn-grad-cyan" :loading="speedSaving" @click="handleSaveSpeed">保存</n-button>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useMessage, useDialog, NButton, NTable, NModal, NForm, NFormItem, NInput, NInputNumber } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import {
  adminGetNodes,
  adminCreateNode,
  adminUpdateNode,
  adminDeleteNode,
  adminStartNode,
  adminStopNode,
  adminSetNodeSpeed
} from '@/api'

const message = useMessage()
const dialog = useDialog()

const nodes = ref([])
const modalVisible = ref(false)
const editing = ref(null)
const saving = ref(false)
const speedVisible = ref(false)
const speedSaving = ref(false)
const speedTarget = ref(null)
const formRef = ref(null)

const emptyForm = () => ({
  name: '',
  address: '',
  port: 7000,
  speed_limit_mbps: 0,
  remark: ''
})

const form = reactive(emptyForm())
const speedForm = reactive({ speed_limit_mbps: 0 })

const rules = {
  name: [{ required: true, message: '请输入节点名称', trigger: ['input', 'blur'] }],
  address: [{ required: true, message: '请输入节点地址', trigger: ['input', 'blur'] }],
  port: [
    {
      validator: (rule, value) => {
        if (value === null || value === undefined || value === '') {
          return new Error('请输入控制端口')
        }
        if (value < 1 || value > 65535) {
          return new Error('端口范围 1-65535')
        }
        return true
      },
      trigger: ['input', 'change', 'blur']
    }
  ]
}

function isOnline(n) {
  return n.status === 'online' || n.status === 1 || n.status === '1' || n.status === 'running'
}

function maskToken(token) {
  if (!token) return '-'
  if (token.length <= 10) return token
  return `${token.slice(0, 6)}...${token.slice(-4)}`
}

async function loadData() {
  try {
    nodes.value = (await adminGetNodes()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, emptyForm())
  modalVisible.value = true
}

function openEdit(n) {
  editing.value = n
  Object.assign(form, {
    name: n.name || '',
    address: n.address || '',
    port: n.port ?? 7000,
    speed_limit_mbps: n.speed_limit_mbps ?? 0,
    remark: n.remark || ''
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
    address: form.address,
    port: form.port,
    speed_limit_mbps: form.speed_limit_mbps,
    remark: form.remark
  }
  try {
    if (editing.value) {
      await adminUpdateNode(editing.value.id, payload)
      message.success('节点已更新')
    } else {
      await adminCreateNode(payload)
      message.success('节点创建成功')
    }
    modalVisible.value = false
    await loadData()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

async function handleStart(n) {
  try {
    await adminStartNode(n.id)
    message.success('节点已启用')
    await loadData()
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function handleStop(n) {
  try {
    await adminStopNode(n.id)
    message.success('节点已停用')
    await loadData()
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function openSpeed(n) {
  speedTarget.value = n
  speedForm.speed_limit_mbps = n.speed_limit_mbps ?? 0
  speedVisible.value = true
}

async function handleSaveSpeed() {
  speedSaving.value = true
  try {
    await adminSetNodeSpeed(speedTarget.value.id, { speed_limit_mbps: speedForm.speed_limit_mbps })
    message.success('限速已更新')
    speedVisible.value = false
    await loadData()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    speedSaving.value = false
  }
}

function handleDelete(n) {
  dialog.warning({
    title: '删除节点',
    content: `确定要删除节点「${n.name}」吗？其下隧道将受影响。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminDeleteNode(n.id)
        message.success('节点已删除')
        await loadData()
      } catch (e) {
        // 错误已由拦截器提示
      }
    }
  })
}

async function copyToken(n) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(n.agent_token)
    } else {
      // 非 HTTPS 环境降级：临时 textarea + execCommand
      const textarea = document.createElement('textarea')
      textarea.value = n.agent_token
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    message.success('Agent Token 已复制')
  } catch (e) {
    message.error('复制失败')
  }
}

onMounted(loadData)
</script>

<style scoped>
.table-wrap {
  padding: 12px 16px;
}

.cell-main {
  min-width: 0;
}

.node-name {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.cell-sub {
  font-size: 12px;
  color: #94a3b8;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.token-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}

.token-text {
  font-size: 12px;
  font-family: Consolas, 'Courier New', monospace;
  color: #64748b;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 16px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

.plan-user-tip {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 14px;
}
</style>
