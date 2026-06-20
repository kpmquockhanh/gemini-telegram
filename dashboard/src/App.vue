<script setup lang="ts">
import { ConfigProvider } from 'ant-design-vue'
import { customTheme } from '@/theme'
import AppLayout from '@/components/AppLayout.vue'
</script>

<template>
  <ConfigProvider :theme="customTheme">
    <AppLayout />
  </ConfigProvider>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Space+Grotesk:wght@500;600;700&display=swap');

/* ── Force light color scheme ── */
html,
body {
  color-scheme: light !important;
}

body {
  margin: 0;
  font-family:
    'Inter',
    -apple-system,
    BlinkMacSystemFont,
    'Segoe UI',
    Roboto,
    'Helvetica Neue',
    Arial,
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background-color: #f8fafc !important;
  color: #0f172a !important;
}

/* ── Scrollbar ── */
::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* ── Global transitions ── */
* {
  transition:
    background-color 0.2s ease,
    color 0.2s ease,
    border-color 0.2s ease;
}

/* ── Layout ── */
.ant-layout {
  background: #f8fafc !important;
}
.ant-layout-sider {
  background: #ffffff !important;
}
.ant-layout-header {
  background: #ffffff !important;
}
.ant-layout-content {
  background: #f8fafc !important;
}

/* ── Table ── */
.ant-table {
  border-radius: 14px !important;
  overflow: hidden;
  background: #ffffff !important;
}
.ant-table-thead > tr > th {
  background: #f8fafc !important;
  font-weight: 600 !important;
  color: #475569 !important;
  border-bottom: 1px solid #e2e8f0 !important;
  font-size: 12px !important;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.ant-table-tbody > tr:hover > td {
  background: #f1f5f9 !important;
}
.ant-table-tbody > tr > td {
  border-bottom: 1px solid #e2e8f0 !important;
  color: #0f172a !important;
  background: #ffffff !important;
}

/* ── Card ── */
.ant-card {
  border-radius: 14px !important;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.04) !important;
  border: 1px solid #e2e8f0 !important;
  background: #ffffff !important;
}

/* ── Modal ── */
.ant-modal-content {
  border-radius: 16px !important;
  overflow: hidden;
  background: #ffffff !important;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.15) !important;
}
.ant-modal-header {
  background: #ffffff !important;
}
.ant-modal-title {
  color: #0f172a !important;
  font-weight: 600 !important;
}

.ant-modal-body {
  background: #ffffff !important;
}

/* ── Button ── */
.ant-btn-primary {
  background: #00d9a5 !important;
  border-color: #00d9a5 !important;
  color: #0f172a !important;
}
.ant-btn-primary:hover {
  background: #00f5b8 !important;
  border-color: #00f5b8 !important;
}
.ant-btn-default {
  background: #ffffff !important;
  border-color: #e2e8f0 !important;
  color: #0f172a !important;
}
.ant-btn-text {
  color: #64748b !important;
}

/* ── Input ── */
.ant-input,
.ant-input-affix-wrapper,
.ant-select-selector,
.ant-picker {
  background: #ffffff !important;
  border-color: #e2e8f0 !important;
  color: #0f172a !important;
}
.ant-input::placeholder,
.ant-input-affix-wrapper::placeholder {
  color: #94a3b8 !important;
}

/* ── Menu ── */
.ant-menu {
  background: transparent !important;
}
.ant-menu-item {
  color: #64748b !important;
}
.ant-menu-item-selected {
  background: #f1f5f9 !important;
  color: #00d9a5 !important;
}

/* ── Select dropdown ── */
.ant-select-dropdown {
  background: #ffffff !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 12px !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08) !important;
}
.ant-select-item {
  color: #0f172a !important;
}
.ant-select-item-option-selected {
  background: #f1f5f9 !important;
  color: #00d9a5 !important;
}
.ant-select-item-option-active {
  background: #f1f5f9 !important;
}

/* ── Form label ── */
.ant-form-item-label > label {
  color: #475569 !important;
  font-weight: 500 !important;
}

/* ── TextArea ── */
.ant-input-textarea {
  background: #ffffff !important;
  border-color: #e2e8f0 !important;
  color: #0f172a !important;
}
.ant-input-textarea::placeholder {
  color: #94a3b8 !important;
}
.ant-input-textarea-show-count {
  color: #94a3b8 !important;
}

/* ── Pagination ── */
.ant-pagination-item {
  background: #ffffff !important;
  border-color: #e2e8f0 !important;
}
.ant-pagination-item-active {
  border-color: #00d9a5 !important;
  color: #00d9a5 !important;
}

/* ── Drawer ── */
.ant-drawer-body {
  background: #ffffff !important;
}
.ant-drawer-header {
  background: #ffffff !important;
  border-bottom: 1px solid #e2e8f0 !important;
}

/* ── Dropdown ── */
.ant-dropdown-menu {
  background: #ffffff !important;
  border-radius: 12px !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08) !important;
}
.ant-dropdown-menu-item {
  color: #0f172a !important;
}
.ant-dropdown-menu-item:hover {
  background: #f1f5f9 !important;
}

/* ── Tag ── */
.ant-tag {
  border-radius: 6px !important;
}

/* ── Switch ── */
.ant-switch-checked {
  background: #00d9a5 !important;
}

/* ── Tabs ── */
.ant-tabs-tab {
  color: #64748b !important;
}
.ant-tabs-tab-active {
  color: #00d9a5 !important;
}
.ant-tabs-ink-bar {
  background: #00d9a5 !important;
}

/* ── Signature: holographic border animation ── */
@keyframes borderRotate {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

@keyframes glowPulse {
  0%,
  100% {
    opacity: 0.4;
  }
  50% {
    opacity: 0.8;
  }
}

.holo-border {
  position: relative;
  border-radius: 16px;
  overflow: hidden;
}
.holo-border::before {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: 17px;
  background: linear-gradient(90deg, #00d9a5, #60a5fa, #00d9a5);
  background-size: 200% 200%;
  animation: borderRotate 4s linear infinite;
  z-index: 0;
  opacity: 0.3;
  pointer-events: none;
}
.holo-border::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 16px;
  background: #ffffff;
  z-index: 1;
}
.holo-border > * {
  position: relative;
  z-index: 2;
}

.holo-border-pulse::before {
  animation:
    borderRotate 4s linear infinite,
    glowPulse 2s ease-in-out infinite;
}
</style>
