<template>
  <div class="glass-card gc">
    <div v-if="title || $slots.extra" class="gc-header">
      <div class="gc-title">
        <svg-icon v-if="icon" :name="icon" :size="17" class="gc-icon" />
        <span>{{ title }}</span>
      </div>
      <div class="gc-extra">
        <slot name="extra" />
      </div>
    </div>
    <div class="gc-body" :style="{ padding }">
      <slot />
    </div>
  </div>
</template>

<script setup>
import SvgIcon from './SvgIcon.vue'

defineProps({
  title: { type: String, default: '' },
  icon: { type: String, default: '' },
  padding: { type: String, default: '20px' }
})
</script>

<style scoped>
.gc {
  margin-bottom: 16px;
}

.gc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px 0;
}

.gc-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

.gc-icon {
  color: #0ea5e9;
}

.gc-extra {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
