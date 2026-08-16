<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>隧道管理</h2>
        <div class="sub">全局隧道列表，可强制下线异常隧道</div>
      </div>
      <n-button size="small" @click="loadData">
        <template #icon><svg-icon name="refresh" :size="14" /></template>
        刷新
      </n-button>
    </div>

    <div class="glass-card table-wrap">
      <n-table :bordered="false" size="small">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户</th>
            <th>名称</th>
            <th>类型</th>
            <th>节点</th>
            <th>本地地址</th>
            <th>远程端口</th>
            <th>状态</th>
            <th style="width: 110px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in tunnels" :key="t.id">
            <td>{{ t.id }}</td>
            <td>{{ t.username || t.user?.username || t.user_id || '-' }}</td>
            <td class="cell-main">
              <div class="t-name">{{ t.name }}</div>
              <div class="cell-sub">{{ t.subdomain || t.custom_domain || '' }}</div>
            </td>
            <td><span class="type-tag">{{ tunnelTypeText(t.type) }}</span></td>
            <td>{{ t.node_name || t.node?.name || '-' }}</td>
            <td>{{ t.local_ip || '-' }}:{{ t.local_port || '-' }}</td>
            <td>{{ t.remote_port || '自动' }}</td>
            <td>
              <status-badge :type="tunnelStatusType(t.online ?? t.status)" :text="tunnelStatusText(t.online ?? t.status)" />
            </td>
            <td>
              <n-button size="tiny" type="error" ghost :disabled="!isRunning(t)" @click="handleOffline(t)">
                强制下线
              </n-button>
            </td>
          </tr>
          <tr v-if="!tunnels.length">
            <td colspan="9">
              <div class="empty-tip">暂无隧道数据</div>
            </td>
          </tr>
        </tbody>
      </n-table>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useMessage, useDialog, NButton, NTable } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { adminGetTunnels, adminOfflineTunnel } from '@/api'
import { tunnelTypeText, tunnelStatusText, tunnelStatusType } from '@/utils/format'

const message = useMessage()
const dialog = useDialog()

const tunnels = ref([])

function isRunning(t) {
  const s = t.online ?? t.status
  return s === 'running' || s === 'online' || s === true || s === 1 || s === '1'
}

async function loadData() {
  try {
    const res = await adminGetTunnels()
    tunnels.value = Array.isArray(res) ? res : res?.items || res?.list || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function handleOffline(t) {
  dialog.warning({
    title: '强制下线',
    content: `确定要强制下线隧道「${t.name}」（ID: ${t.id}）吗？`,
    positiveText: '强制下线',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminOfflineTunnel(t.id)
        message.success('已强制下线')
        await loadData()
      } catch (e) {
        // 错误已由拦截器提示
      }
    }
  })
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

.t-name {
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
</style>
