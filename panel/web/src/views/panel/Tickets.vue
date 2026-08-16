<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>工单中心</h2>
        <div class="sub">遇到问题？提交工单，客服会尽快处理</div>
      </div>
      <n-button class="btn-grad-orange" @click="openCreate">
        <template #icon><svg-icon name="plus" :size="16" /></template>
        新建工单
      </n-button>
    </div>

    <div class="glass-card list-card">
      <div v-if="list.length" class="tk-list">
        <div v-for="item in list" :key="item.id" class="tk-item" @click="openDetail(item)">
          <div class="tk-left">
            <span class="tk-title">{{ item.title }}</span>
            <span class="tk-preview">{{ item.content }}</span>
          </div>
          <div class="tk-right">
            <status-badge :type="isClosed(item) ? 'default' : 'info'" :text="isClosed(item) ? '已关闭' : '处理中'" />
            <span class="tk-date">{{ formatDate(item.created_at, 'YYYY-MM-DD') }}</span>
          </div>
        </div>
      </div>
      <div v-else class="empty-tip">暂无工单，有需要就创建一条吧</div>
    </div>

    <n-modal v-model:show="createVisible" preset="card" title="新建工单" style="width: 560px; max-width: 94vw">
      <n-form ref="createRef" :model="createForm" :rules="createRules" label-placement="top">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="createForm.title" placeholder="请简要描述问题" clearable />
        </n-form-item>
        <n-form-item label="问题描述" path="content">
          <n-input v-model:value="createForm.content" type="textarea" :rows="5" placeholder="请详细描述你遇到的问题，包括节点、隧道名称等" />
        </n-form-item>
      </n-form>
      <div class="modal-actions">
        <n-button @click="createVisible = false">取消</n-button>
        <n-button class="btn-grad-cyan" :loading="creating" @click="handleCreate">提交工单</n-button>
      </div>
    </n-modal>

    <n-modal v-model:show="detailVisible" preset="card" :title="detail.title || '工单详情'" style="width: 640px; max-width: 94vw">
      <div class="detail-meta">
        <status-badge :type="isClosed(detail) ? 'default' : 'info'" :text="isClosed(detail) ? '已关闭' : '处理中'" />
        <span>创建于 {{ formatDate(detail.created_at) }}</span>
      </div>
      <div class="msg user-msg">
        <div class="msg-label">问题描述</div>
        <div class="msg-content">{{ detail.content }}</div>
      </div>
      <div v-if="detail.admin_reply" class="msg admin-msg">
        <div class="msg-label">官方回复</div>
        <div class="msg-content">{{ detail.admin_reply }}</div>
      </div>
      <template v-if="!isClosed(detail)">
        <n-input v-model:value="replyText" type="textarea" :rows="3" placeholder="补充说明或回复" />
        <div class="modal-actions">
          <n-button type="error" ghost :loading="closing" @click="handleClose">关闭工单</n-button>
          <n-button class="btn-grad-cyan" :loading="replying" @click="handleReply">
            <template #icon><svg-icon name="send" :size="14" /></template>
            发送回复
          </n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useMessage, NButton, NModal, NForm, NFormItem, NInput } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { getTickets, getTicketDetail, createTicket, replyTicket, closeTicket } from '@/api'
import { formatDate } from '@/utils/format'

const message = useMessage()

const list = ref([])
const createVisible = ref(false)
const detailVisible = ref(false)
const creating = ref(false)
const replying = ref(false)
const closing = ref(false)
const replyText = ref('')
const createRef = ref(null)
const detail = reactive({})

const createForm = reactive({ title: '', content: '' })
const createRules = {
  title: [{ required: true, message: '请输入标题', trigger: ['input', 'blur'] }],
  content: [{ required: true, message: '请输入问题描述', trigger: ['input', 'blur'] }]
}

function isClosed(item) {
  const s = item.status
  return s === 'closed' || s === 'close' || s === 0 || s === '0'
}

async function loadList() {
  try {
    list.value = (await getTickets()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function openCreate() {
  createForm.title = ''
  createForm.content = ''
  createVisible.value = true
}

async function handleCreate() {
  try {
    await createRef.value?.validate()
  } catch (e) {
    return
  }
  creating.value = true
  try {
    await createTicket({ title: createForm.title, content: createForm.content })
    message.success('工单已提交')
    createVisible.value = false
    await loadList()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    creating.value = false
  }
}

async function openDetail(item) {
  detailVisible.value = true
  Object.assign(detail, { id: item.id, title: item.title, content: item.content, status: item.status, admin_reply: item.admin_reply, created_at: item.created_at })
  replyText.value = ''
  try {
    const data = await getTicketDetail(item.id)
    if (data) Object.assign(detail, data)
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function handleReply() {
  if (!replyText.value.trim()) {
    message.warning('请输入回复内容')
    return
  }
  replying.value = true
  try {
    await replyTicket(detail.id, { content: replyText.value })
    message.success('回复成功')
    replyText.value = ''
    await openDetail({ ...detail })
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    replying.value = false
  }
}

async function handleClose() {
  closing.value = true
  try {
    await closeTicket(detail.id)
    message.success('工单已关闭')
    detailVisible.value = false
    await loadList()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    closing.value = false
  }
}

onMounted(loadList)
</script>

<style scoped>
.list-card {
  padding: 12px 20px;
}

.tk-list {
  display: flex;
  flex-direction: column;
}

.tk-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.6);
  cursor: pointer;
  border-radius: 8px;
}

.tk-item:last-child {
  border-bottom: none;
}

.tk-item:hover {
  background: rgba(14, 165, 233, 0.06);
}

.tk-left {
  min-width: 0;
  flex: 1;
}

.tk-title {
  font-size: 14px;
  color: #1e293b;
  font-weight: 500;
  display: block;
}

.tk-preview {
  display: block;
  font-size: 12px;
  color: #94a3b8;
  margin-top: 3px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tk-right {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-shrink: 0;
}

.tk-date {
  font-size: 12px;
  color: #94a3b8;
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 14px;
}

.msg {
  margin-bottom: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 13px;
}

.msg-label {
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 6px;
}

.user-msg {
  background: rgba(14, 165, 233, 0.08);
  border: 1px solid rgba(14, 165, 233, 0.18);
}

.user-msg .msg-label {
  color: #0284c7;
}

.admin-msg {
  background: rgba(245, 158, 11, 0.08);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.admin-msg .msg-label {
  color: #d97706;
}

.msg-content {
  color: #334155;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 14px;
}
</style>
