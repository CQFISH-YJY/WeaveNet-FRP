<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>套餐配置</h2>
        <div class="sub">配置四档会员套餐的限速与额度</div>
      </div>
      <n-button size="small" @click="loadData">
        <template #icon><svg-icon name="refresh" :size="14" /></template>
        刷新
      </n-button>
    </div>

    <div class="plan-grid">
      <div v-for="p in sortedPlans" :key="p.id" class="glass-card plan-card">
        <div class="plan-head">
          <span class="plan-name">{{ p.name }}</span>
          <span v-if="p.sort === 0 || p.id === 1" class="free-tag">免费</span>
        </div>
        <div class="plan-speed">
          <span class="speed-num">{{ p.speed_limit_mbps ?? '-' }}</span>
          <span class="speed-unit">Mbps</span>
        </div>
        <div class="plan-meta">限速</div>
        <div class="plan-items">
          <div class="plan-item">
            <span>隧道数量</span>
            <span class="item-val">{{ p.tunnel_limit ?? '-' }} 条</span>
          </div>
          <div class="plan-item">
            <span>流量</span>
            <span class="item-val">不限</span>
          </div>
        </div>
        <n-button class="btn-grad-cyan edit-btn" block @click="openEdit(p)">
          <template #icon><svg-icon name="edit" :size="14" /></template>
          编辑套餐
        </n-button>
      </div>
    </div>

    <n-modal v-model:show="modalVisible" preset="card" title="编辑套餐" style="width: 500px; max-width: 94vw">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top">
        <n-form-item label="套餐名称" path="name">
          <n-input v-model:value="form.name" clearable />
        </n-form-item>
        <div class="form-grid">
          <n-form-item label="限速 (Mbps)" path="speed_limit_mbps">
            <n-input-number v-model:value="form.speed_limit_mbps" :min="0" style="width: 100%" />
          </n-form-item>
          <n-form-item label="排序" path="sort">
            <n-input-number v-model:value="form.sort" :min="0" style="width: 100%" />
          </n-form-item>
          <n-form-item label="隧道数量" path="tunnel_limit">
            <n-input-number v-model:value="form.tunnel_limit" :min="0" style="width: 100%" />
          </n-form-item>
        </div>
      </n-form>
      <div class="modal-actions">
        <n-button @click="modalVisible = false">取消</n-button>
        <n-button class="btn-grad-cyan" :loading="saving" @click="handleSave">保存</n-button>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useMessage, NButton, NModal, NForm, NFormItem, NInput, NInputNumber } from 'naive-ui'
import SvgIcon from '@/components/SvgIcon.vue'
import { adminGetPlans, adminUpdatePlan } from '@/api'

const message = useMessage()

const plans = ref([])
const modalVisible = ref(false)
const saving = ref(false)
const editing = ref(null)
const formRef = ref(null)

const form = reactive({
  name: '',
  speed_limit_mbps: 0,
  tunnel_limit: 0,
  sort: 0
})

const rules = {
  name: [{ required: true, message: '请输入套餐名称', trigger: ['input', 'blur'] }]
}

const sortedPlans = computed(() => [...plans.value].sort((a, b) => (a.sort ?? a.id ?? 0) - (b.sort ?? b.id ?? 0)))

async function loadData() {
  try {
    plans.value = (await adminGetPlans()) || []
  } catch (e) {
    // 错误已由拦截器提示
  }
}

function openEdit(p) {
  editing.value = p
  Object.assign(form, {
    name: p.name || '',
    speed_limit_mbps: p.speed_limit_mbps ?? 0,
    tunnel_limit: p.tunnel_limit ?? 0,
    sort: p.sort ?? 0
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
  try {
    await adminUpdatePlan(editing.value.id, {
      name: form.name,
      speed_limit_mbps: form.speed_limit_mbps,
      tunnel_limit: form.tunnel_limit,
      sort: form.sort
    })
    message.success('套餐已更新')
    modalVisible.value = false
    await loadData()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.plan-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.plan-card {
  padding: 24px 20px;
}

.plan-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.plan-name {
  font-size: 17px;
  font-weight: 700;
  color: #0f172a;
}

.free-tag {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  color: #64748b;
  background: rgba(100, 116, 139, 0.12);
  border: 1px solid rgba(100, 116, 139, 0.25);
}

.plan-speed {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.speed-num {
  font-size: 36px;
  font-weight: 800;
  line-height: 1.2;
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.speed-unit {
  font-size: 13px;
  color: #64748b;
}

.plan-meta {
  font-size: 12px;
  color: #94a3b8;
  margin: 4px 0 16px;
}

.plan-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 18px;
}

.plan-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.6);
  font-size: 13px;
  color: #475569;
}

.item-val {
  font-weight: 600;
  color: #0ea5e9;
}

.edit-btn {
  width: 100%;
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

@media (max-width: 1100px) {
  .plan-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
