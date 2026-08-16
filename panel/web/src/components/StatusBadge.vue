<template>
  <span class="status-badge" :class="`sb-${type}`">
    <span class="sb-dot"></span>
    <span>{{ text }}</span>
  </span>
</template>

<script setup>
defineProps({
  text: { type: String, default: '' },
  type: { type: String, default: 'default' }
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 1.6;
  white-space: nowrap;
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.7);
}

.sb-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  flex-shrink: 0;
}

.sb-success {
  color: #059669;
  background: rgba(16, 185, 129, 0.12);
}

.sb-warning {
  color: #d97706;
  background: rgba(245, 158, 11, 0.14);
}

.sb-error {
  color: #dc2626;
  background: rgba(239, 68, 68, 0.12);
}

.sb-info {
  color: #0284c7;
  background: rgba(14, 165, 233, 0.12);
}

.sb-default {
  color: #64748b;
}
</style>
