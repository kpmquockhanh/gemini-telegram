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

body {
  margin: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto,
    'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background-color: #F8FAFC;
  color: #0F172A;
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
  background: #CBD5E1;
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover {
  background: #94A3B8;
}

/* ── Global transitions ── */
* {
  transition: background-color 0.2s ease, color 0.2s ease,
    border-color 0.2s ease;
}

/* ── Table ── */
.ant-table {
  border-radius: 14px !important;
  overflow: hidden;
  background: #FFFFFF !important;
}
.ant-table-thead > tr > th {
  background: #F8FAFC !important;
  font-weight: 600 !important;
  color: #475569 !important;
  border-bottom: 1px solid #E2E8F0 !important;
  font-size: 12px !important;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.ant-table-tbody > tr:hover > td {
  background: #F1F5F9 !important;
}
.ant-table-tbody > tr > td {
  border-bottom: 1px solid #E2E8F0 !important;
  color: #0F172A !important;
}

/* ── Card ── */
.ant-card {
  border-radius: 14px !important;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.04) !important;
  border: 1px solid #E2E8F0 !important;
  background: #FFFFFF !important;
}

/* ── Modal ── */
.ant-modal-content {
  border-radius: 16px !important;
  overflow: hidden;
  background: #FFFFFF !important;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.15) !important;
}
.ant-modal-header {
  border-bottom: 1px solid #E2E8F0 !important;
  padding: 20px 24px !important;
  background: #FFFFFF !important;
}
.ant-modal-title {
  color: #0F172A !important;
  font-weight: 600 !important;
}
.ant-modal-footer {
  border-top: 1px solid #E2E8F0 !important;
}

/* ── Button ── */
.ant-btn {
  border-radius: 10px !important;
  font-weight: 500 !important;
}
.ant-btn-primary {
  background: #00D9A5 !important;
  border-color: #00D9A5 !important;
  color: #0F172A !important;
}
.ant-btn-primary:hover {
  background: #00F5B8 !important;
  border-color: #00F5B8 !important;
}

/* ── Input ── */
.ant-input,
.ant-input-affix-wrapper,
.ant-select-selector,
.ant-picker {
  border-radius: 10px !important;
  background: #FFFFFF !important;
  border-color: #E2E8F0 !important;
  color: #0F172A !important;
}
.ant-input::placeholder,
.ant-input-affix-wrapper::placeholder {
  color: #94A3B8 !important;
}

/* ── Menu ── */
.ant-menu {
  background: transparent !important;
}
.ant-menu-item {
  border-radius: 10px !important;
  margin: 4px 8px !important;
  color: #64748B !important;
}

/* ── Select dropdown ── */
.ant-select-dropdown {
  background: #FFFFFF !important;
  border: 1px solid #E2E8F0 !important;
  border-radius: 12px !important;
}
.ant-select-item {
  color: #0F172A !important;
}
.ant-select-item-option-selected {
  background: #F1F5F9 !important;
  color: #00D9A5 !important;
}
.ant-select-item-option-active {
  background: #F1F5F9 !important;
}

/* ── Form label ── */
.ant-form-item-label > label {
  color: #475569 !important;
  font-weight: 500 !important;
}

/* ── TextArea ── */
.ant-input-textarea {
  background: #FFFFFF !important;
  border-color: #E2E8F0 !important;
  color: #0F172A !important;
}
.ant-input-textarea::placeholder {
  color: #94A3B8 !important;
}
.ant-input-textarea-show-count {
  color: #94A3B8 !important;
}

/* ── Signature: holographic border animation ── */
@keyframes borderRotate {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

@keyframes glowPulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
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
  background: linear-gradient(90deg, #00D9A5, #60A5FA, #00D9A5);
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
  background: #FFFFFF;
  z-index: 1;
}
.holo-border > * {
  position: relative;
  z-index: 2;
}

.holo-border-pulse::before {
  animation: borderRotate 4s linear infinite, glowPulse 2s ease-in-out infinite;
}
</style>
