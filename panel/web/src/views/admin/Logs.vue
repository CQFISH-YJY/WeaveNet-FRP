<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>操作日志</h2>
        <div class="sub">管理员后台操作审计记录</div>
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
            <th>管理员</th>
            <th>操作</th>
            <th>目标类型</th>
            <th>目标 ID</th>
            <th>详情</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id">
            <td>{{ log.id }}</td>
            <td>{{ log.admin_name || log.admin_id || '-' }}</td>
            <td>
              <span class="action-tag">{{ log.action || '-' }}</span>
            </td>
            <td>{{ log.target_type || '-' }}</td>
            <td>{{ log.target_id ?? '-' }}</td>
            <td class="cell-sub detail-cell" :title="log.detail">{{ log.detail || '-' }}</td>
            <td>{{ formatDate(log.created_at) }}</td>
          </tr>
          <tr v-if="!logs.length">
            <td colspan="7">
              <div class="empty-tip">暂无操作日志</div>
            </td>
          </tr>
        </tbody>
      </n-table>
      <div class="pager">
        <n-pagination
          v-model:page="query.page"
          :page-size="query.page_size"
          :item-count="total"
          @update:page="loadData"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { NButton, NTable, NPagination } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { adminGetOperationLogs } from '@/api'
import { formatDate } from '@/utils/format'

const logs = ref([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 15 })

async function loadData() {
  try {
    const res = await adminGetOperationLogs({
      page: query.page,
      page_size: query.page_size
    })
    if (Array.isArray(res)) {
      logs.value = res
      total.value = res.length
    } else {
      logs.value = res?.items || res?.list || []
      total.value = res?.total || logs.value.length
    }
  } catch (e) {
    // 错误已由拦截器提示
  }
}

onMounted(loadData)
</script>

<style scoped>
.table-wrap {
  padding: 12px 16px;
}

.cell-sub {
  font-size: 12px;
  color: #64748b;
}

.detail-cell {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #0284c7;
  background: rgba(14, 165, 233, 0.12);
  border: 1px solid rgba(14, 165, 233, 0.2);
}

.pager {
  display: flex;
  justify-content: flex-end;
  padding: 14px 4px 4px;
}
</style>
