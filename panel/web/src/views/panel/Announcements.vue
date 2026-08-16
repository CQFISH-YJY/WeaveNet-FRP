<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>平台公告</h2>
        <div class="sub">了解平台最新动态与维护通知</div>
      </div>
    </div>

    <div class="glass-card list-card">
      <div v-if="list.length" class="ann-list">
        <div v-for="item in list" :key="item.id" class="ann-item" @click="openDetail(item)">
          <div class="ann-left">
            <span class="ann-tag">公告</span>
            <span class="ann-title">{{ item.title }}</span>
          </div>
          <div class="ann-right">
            <span class="ann-author">{{ item.author || '官方' }}</span>
            <span class="ann-date">{{ formatDate(item.created_at, 'YYYY-MM-DD') }}</span>
            <svg-icon name="chevronDown" :size="14" class="ann-arrow" />
          </div>
        </div>
      </div>
      <div v-else class="empty-tip">暂无公告</div>
    </div>

    <n-modal v-model:show="detailVisible" preset="card" :title="detail.title || '公告详情'" style="width: 640px; max-width: 94vw">
      <div class="detail-meta">
        <span>发布人：{{ detail.author || '官方' }}</span>
        <span>发布于：{{ formatDate(detail.created_at) }}</span>
      </div>
      <div class="detail-content">{{ detail.content || detail.body || '' }}</div>
    </n-modal>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { NModal } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { getAnnouncements, getAnnouncementDetail } from '@/api'
import { formatDate } from '@/utils/format'

const list = ref([])
const detailVisible = ref(false)
const detail = reactive({})

async function loadList() {
  try {
    list.value = (await getAnnouncements()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function openDetail(item) {
  detailVisible.value = true
  Object.assign(detail, { title: item.title, author: item.author, content: item.content, created_at: item.created_at })
  try {
    const data = await getAnnouncementDetail(item.id)
    if (data) Object.assign(detail, data)
  } catch (e) {
    // 列表已有内容时详情失败不阻塞
  }
}

onMounted(loadList)
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
  cursor: pointer;
  border-radius: 8px;
}

.ann-item:last-child {
  border-bottom: none;
}

.ann-item:hover {
  background: rgba(14, 165, 233, 0.06);
}

.ann-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.ann-tag {
  flex-shrink: 0;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #b45309;
  background: rgba(245, 158, 11, 0.14);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.ann-title {
  font-size: 14px;
  color: #1e293b;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ann-right {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-shrink: 0;
  font-size: 12px;
  color: #94a3b8;
}

.ann-arrow {
  transition: transform 0.2s;
  color: #cbd5e1;
}

.ann-item:hover .ann-arrow {
  transform: rotate(180deg);
  color: #0ea5e9;
}

.detail-meta {
  display: flex;
  gap: 20px;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 14px;
}

.detail-content {
  font-size: 14px;
  line-height: 1.9;
  color: #334155;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 50vh;
  overflow: auto;
}
</style>
