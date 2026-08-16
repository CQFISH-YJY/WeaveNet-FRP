<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>公告管理</h2>
        <div class="sub">发布与维护平台公告</div>
      </div>
      <n-button class="btn-grad-orange" @click="openCreate">
        <template #icon><svg-icon name="plus" :size="16" /></template>
        发布公告
      </n-button>
    </div>

    <div class="glass-card list-card">
      <div v-if="list.length" class="ann-list">
        <div v-for="item in list" :key="item.id" class="ann-item">
          <div class="ann-left">
            <span class="ann-title">{{ item.title }}</span>
            <span class="ann-author">发布人：{{ item.author || '-' }}</span>
          </div>
          <div class="ann-mid">{{ formatDate(item.created_at) }}</div>
          <div class="ann-right">
            <status-badge :type="isOnline(item) ? 'success' : 'default'" :text="isOnline(item) ? '已发布' : '已下线'" />
            <div class="table-actions">
              <n-button size="tiny" @click="openEdit(item)">
                <template #icon><svg-icon name="edit" :size="13" /></template>
                编辑
              </n-button>
              <n-button v-if="isOnline(item)" size="tiny" type="error" ghost @click="handleOffline(item)">
                下线
              </n-button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-tip">暂无公告</div>
    </div>

    <n-modal v-model:show="modalVisible" preset="card" :title="editing ? '编辑公告' : '发布公告'" style="width: 620px; max-width: 94vw">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="form.title" placeholder="公告标题" clearable />
        </n-form-item>
        <n-form-item label="发布人" path="author">
          <n-input v-model:value="form.author" placeholder="例如：WeaveNet 官方" clearable />
        </n-form-item>
        <n-form-item label="内容" path="content">
          <n-input v-model:value="form.content" type="textarea" :rows="6" placeholder="公告正文内容" />
        </n-form-item>
      </n-form>
      <div class="modal-actions">
        <n-button @click="modalVisible = false">取消</n-button>
        <n-button class="btn-grad-cyan" :loading="saving" @click="handleSave">
          {{ editing ? '保存修改' : '发布' }}
        </n-button>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useMessage, useDialog, NButton, NModal, NForm, NFormItem, NInput } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { adminGetAnnouncements, adminCreateAnnouncement, adminUpdateAnnouncement, adminOfflineAnnouncement } from '@/api'
import { formatDate } from '@/utils/format'
import { useUserStore } from '@/store/user'

const message = useMessage()
const dialog = useDialog()
const userStore = useUserStore()

const list = ref([])
const modalVisible = ref(false)
const saving = ref(false)
const editing = ref(null)
const formRef = ref(null)

const form = reactive({ title: '', author: '', content: '' })

const rules = {
  title: [{ required: true, message: '请输入公告标题', trigger: ['input', 'blur'] }],
  author: [{ required: true, message: '请输入发布人', trigger: ['input', 'blur'] }],
  content: [{ required: true, message: '请输入公告内容', trigger: ['input', 'blur'] }]
}

function isOnline(item) {
  const s = item.status
  return s === 1 || s === '1' || s === 'active' || s === 'online' || s === 'published'
}

async function loadData() {
  try {
    list.value = (await adminGetAnnouncements()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, {
    title: '',
    author: userStore.displayName || '官方',
    content: ''
  })
  modalVisible.value = true
}

function openEdit(item) {
  editing.value = item
  Object.assign(form, {
    title: item.title || '',
    author: item.author || '',
    content: item.content || ''
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
  const payload = { title: form.title, author: form.author, content: form.content }
  try {
    if (editing.value) {
      await adminUpdateAnnouncement(editing.value.id, payload)
      message.success('公告已更新')
    } else {
      await adminCreateAnnouncement(payload)
      message.success('公告发布成功')
    }
    modalVisible.value = false
    await loadData()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

function handleOffline(item) {
  dialog.warning({
    title: '公告下线',
    content: `确定要下线公告「${item.title}」吗？下线后用户将不可见。`,
    positiveText: '确认下线',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminOfflineAnnouncement(item.id)
        message.success('公告已下线')
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
.list-card {
  padding: 12px 20px;
}

.ann-list {
  display: flex;
  flex-direction: column;
}

.ann-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.6);
}

.ann-item:last-child {
  border-bottom: none;
}

.ann-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
}

.ann-title {
  font-size: 14px;
  color: #1e293b;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ann-author {
  font-size: 12px;
  color: #94a3b8;
}

.ann-mid {
  font-size: 12px;
  color: #94a3b8;
  flex-shrink: 0;
}

.ann-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}
</style>
