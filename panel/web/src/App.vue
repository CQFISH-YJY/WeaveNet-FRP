<template>
  <n-config-provider :theme-overrides="themeOverrides" :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider placement="top">
      <n-dialog-provider>
        <n-notification-provider>
          <router-view />
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import {
  zhCN,
  dateZhCN,
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider
} from 'naive-ui'

const themeOverrides = {
  common: {
    primaryColor: '#0ea5e9',
    primaryColorHover: '#06b6d4',
    primaryColorPressed: '#0284c7',
    primaryColorSuppl: '#06b6d4',
    infoColor: '#0ea5e9',
    infoColorHover: '#06b6d4',
    successColor: '#10b981',
    warningColor: '#f59e0b',
    errorColor: '#ef4444',
    borderRadius: '10px',
    fontFamily:
      "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif"
  },
  Card: {
    borderRadius: '16px'
  },
  Button: {
    borderRadiusMedium: '10px',
    borderRadiusLarge: '12px',
    fontWeight: '500'
  },
  Input: {
    borderRadius: '10px'
  },
  Select: {
    borderRadius: '10px'
  },
  DatePicker: {
    borderRadius: '10px'
  },
  Menu: {
    borderRadius: '10px'
  }
}
</script>

<style>
* {
  box-sizing: border-box;
}

html,
body,
#app {
  height: 100%;
  margin: 0;
  padding: 0;
}

body {
  font-family:
    -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  color: #1e293b;
  background: linear-gradient(160deg, #f0f7ff 0%, #e6f4f4 55%, #eef7f0 100%);
  background-attachment: fixed;
  -webkit-font-smoothing: antialiased;
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-thumb {
  background: rgba(14, 165, 233, 0.25);
  border-radius: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

.glass-card {
  position: relative;
  background: rgba(255, 255, 255, 0.55);
  -webkit-backdrop-filter: blur(16px);
  backdrop-filter: blur(16px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.65);
  overflow: hidden;
}

.glass-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 16px;
  padding: 1px;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.95) 0%,
    rgba(255, 255, 255, 0.25) 35%,
    rgba(14, 165, 233, 0.35) 100%
  );
  -webkit-mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  mask-composite: exclude;
  pointer-events: none;
}

.glass-card-body {
  position: relative;
  z-index: 1;
}

.btn-grad-cyan,
.btn-grad-orange {
  border: none !important;
  color: #fff !important;
}

.btn-grad-cyan.n-button:not(.n-button--disabled) {
  background: linear-gradient(135deg, #0ea5e9, #06b6d4) !important;
}

.btn-grad-cyan.n-button:not(.n-button--disabled):hover {
  background: linear-gradient(135deg, #06b6d4, #0891b2) !important;
}

.btn-grad-orange.n-button:not(.n-button--disabled) {
  background: linear-gradient(135deg, #f59e0b, #f97316) !important;
}

.btn-grad-orange.n-button:not(.n-button--disabled):hover {
  background: linear-gradient(135deg, #f97316, #ea580c) !important;
}

.page-container {
  padding: 20px 24px 32px;
}

.page-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}

.page-title h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #0f172a;
}

.page-title .sub {
  margin-top: 4px;
  font-size: 13px;
  color: #64748b;
}

.stat-value {
  font-size: 26px;
  font-weight: 700;
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.stat-label {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 6px;
}

.gradient-text {
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.text-muted {
  color: #64748b;
}

.text-sm {
  font-size: 12px;
}

.text-warn {
  color: #f59e0b;
}

.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.auth-card {
  width: 100%;
  max-width: 430px;
  padding: 36px 32px 24px;
}

.auth-brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 10px;
}

.auth-brand-name {
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, #0ea5e9, #06b6d4);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.auth-sub {
  text-align: center;
  color: #64748b;
  font-size: 13px;
  margin-bottom: 24px;
}

.auth-links {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 2px 0 18px;
  font-size: 13px;
}

.auth-link {
  color: #0ea5e9;
  cursor: pointer;
  text-decoration: none;
}

.auth-link:hover {
  color: #06b6d4;
  text-decoration: underline;
}

.auth-submit {
  margin-top: 6px;
}

.auth-foot {
  margin-top: 20px;
  text-align: center;
  font-size: 12px;
  color: #94a3b8;
}

.auth-tip {
  margin-top: 14px;
  padding: 10px 12px;
  border-radius: 10px;
  font-size: 12px;
  line-height: 1.7;
  background: rgba(14, 165, 233, 0.1);
  color: #0369a1;
  border: 1px solid rgba(14, 165, 233, 0.2);
}

.menu-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 14px;
  color: #475569;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  border: 1px solid transparent;
}

.menu-item:hover {
  background: rgba(14, 165, 233, 0.08);
  color: #0ea5e9;
}

.menu-item.active {
  background: rgba(14, 165, 233, 0.14);
  color: #0284c7;
  border-color: rgba(14, 165, 233, 0.22);
  font-weight: 600;
}

.menu-item .mi-icon {
  display: inline-flex;
  flex-shrink: 0;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.table-actions .n-button {
  padding: 0 8px;
}

.code-block {
  background: rgba(15, 23, 42, 0.9);
  color: #a5f3fc;
  border-radius: 10px;
  padding: 14px;
  font-size: 12px;
  line-height: 1.7;
  font-family: 'JetBrains Mono', Consolas, 'Courier New', monospace;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 320px;
  overflow: auto;
}

.empty-tip {
  padding: 40px 0;
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
}

</style>
