<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>用户管理</h2>
        <div class="sub">查看与管理平台注册用户</div>
      </div>
    </div>

    <div class="glass-card filter-bar">
      <n-input v-model:value="query.keyword" placeholder="搜索用户名或邮箱" clearable style="width: 240px" @keyup.enter="handleSearch">
        <template #prefix><svg-icon name="search" :size="15" /></template>
      </n-input>
      <n-select
        v-model:value="query.status"
        :options="statusOptions"
        style="width: 140px"
        placeholder="全部状态"
      />
      <n-button class="btn-grad-cyan" @click="handleSearch">搜索</n-button>
      <n-button @click="handleReset">重置</n-button>
    </div>

    <div class="glass-card table-wrap">
      <n-table :bordered="false" size="small">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>邮箱</th>
            <th>状态</th>
            <th>积分</th>
            <th>套餐</th>
            <th>注册时间</th>
            <th style="width: 240px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td>
            <td class="cell-main">
              <div class="user-name">{{ u.username }}</div>
              <div class="cell-sub">{{ u.email_verified ? '邮箱已验证' : '未验证邮箱' }}</div>
            </td>
            <td>{{ u.email || '-' }}</td>
            <td>
              <status-badge :type="isBanned(u) ? 'error' : 'success'" :text="isBanned(u) ? '已封禁' : '正常'" />
            </td>
            <td>{{ u.points ?? 0 }}</td>
            <td>
              <span class="plan-chip">{{ u.plan_name || '免费版' }}</span>
              <div v-if="u.plan_expires_at" class="cell-sub">{{ formatDate(u.plan_expires_at, 'YYYY-MM-DD') }} 到期</div>
            </td>
            <td>{{ formatDate(u.created_at, 'YYYY-MM-DD') }}</td>
            <td>
              <div class="table-actions">
                <n-button v-if="!isBanned(u)" size="tiny" type="error" ghost @click="handleBan(u)">封禁</n-button>
                <n-button v-else size="tiny" type="success" ghost @click="handleUnban(u)">解封</n-button>
                <n-button size="tiny" @click="handleResetPwd(u)">重置密码</n-button>
                <n-button size="tiny" class="plan-btn" @click="openPlanModal(u)">调整套餐</n-button>
              </div>
            </td>
          </tr>
          <tr v-if="!users.length">
            <td colspan="8">
              <div class="empty-tip">暂无匹配的用户</div>
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

    <n-modal v-model:show="planModalVisible" preset="card" title="调整套餐" style="width: 460px; max-width: 94vw">
      <div class="plan-user-tip">用户：{{ planTarget?.username }}（{{ planTarget?.email }}）</div>
      <n-select v-model:value="planForm.plan_id" :options="planOptions" placeholder="选择目标套餐" />
      <div class="modal-actions">
        <n-button @click="planModalVisible = false">取消</n-button>
        <n-button class="btn-grad-cyan" :loading="planSaving" @click="handleSetPlan">保存</n-button>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useMessage, useDialog, NButton, NTable, NInput, NSelect, NPagination, NModal } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { adminGetUsers, adminBanUser, adminUnbanUser, adminResetUserPassword, adminSetUserPlan, adminGetPlans } from '@/api'
import { formatDate } from '@/utils/format'

const message = useMessage()
const dialog = useDialog()

const users = ref([])
const plans = ref([])
const total = ref(0)
const planModalVisible = ref(false)
const planSaving = ref(false)
const planTarget = ref(null)

const query = reactive({
  keyword: '',
  status: null,
  page: 1,
  page_size: 10
})

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '正常', value: 'active' },
  { label: '已封禁', value: 'banned' }
]

const planOptions = computed(() => plans.value.map((p) => ({ label: p.name, value: p.id })))

const planForm = reactive({ plan_id: null })

function isBanned(u) {
  return u.status === 0 || u.status === '0' || u.status === 'banned' || u.status === false
}

async function loadPlans() {
  try {
    plans.value = (await adminGetPlans()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function loadData() {
  try {
    const params = {
      page: query.page,
      page_size: query.page_size
    }
    if (query.keyword) params.keyword = query.keyword
    if (query.status !== null && query.status !== undefined) params.status = query.status
    const res = await adminGetUsers(params)
    if (Array.isArray(res)) {
      users.value = res
      total.value = res.length
    } else {
      users.value = res?.items || res?.list || []
      total.value = res?.total || users.value.length
    }
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function handleSearch() {
  query.page = 1
  loadData()
}

function handleReset() {
  query.keyword = ''
  query.status = null
  query.page = 1
  loadData()
}

function handleBan(u) {
  dialog.warning({
    title: '封禁用户',
    content: `确定要封禁用户「${u.username}」吗？封禁后该用户将无法登录。`,
    positiveText: '确认封禁',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminBanUser(u.id)
        message.success('已封禁该用户')
        loadData()
      } catch (e) {
        // 错误已由拦截器提示
      }
    }
  })
}

function handleUnban(u) {
  dialog.info({
    title: '解除封禁',
    content: `确定要解除用户「${u.username}」的封禁吗？`,
    positiveText: '确认解封',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminUnbanUser(u.id)
        message.success('已解除封禁')
        loadData()
      } catch (e) {
        // 错误已由拦截器提示
      }
    }
  })
}

function handleResetPwd(u) {
  dialog.warning({
    title: '重置密码',
    content: `将重置用户「${u.username}」的密码为随机密码，确定继续吗？`,
    positiveText: '确认重置',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const data = await adminResetUserPassword(u.id)
        message.success(`密码已重置${data?.password ? `：${data.password}` : ''}`)
      } catch (e) {
        // 错误已由拦截器提示
      }
    }
  })
}

function openPlanModal(u) {
  planTarget.value = u
  planForm.plan_id = u.plan_id ?? null
  planModalVisible.value = true
}

async function handleSetPlan() {
  if (!planForm.plan_id) {
    message.warning('请选择目标套餐')
    return
  }
  planSaving.value = true
  try {
    await adminSetUserPlan(planTarget.value.id, { plan_id: planForm.plan_id })
    message.success('套餐已调整')
    planModalVisible.value = false
    loadData()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    planSaving.value = false
  }
}

onMounted(() => {
  loadData()
  loadPlans()
})
</script>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  margin-bottom: 16px;
}

.table-wrap {
  padding: 12px 16px;
}

.cell-main {
  min-width: 0;
}

.user-name {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.cell-sub {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

.plan-chip {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #b45309;
  background: rgba(245, 158, 11, 0.14);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.plan-btn {
  color: #0284c7 !important;
  border-color: rgba(14, 165, 233, 0.35) !important;
}

.pager {
  display: flex;
  justify-content: flex-end;
  padding: 14px 4px 4px;
}

.plan-user-tip {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 14px;
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(14, 165, 233, 0.08);
  border: 1px solid rgba(14, 165, 233, 0.16);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
}
</style>
