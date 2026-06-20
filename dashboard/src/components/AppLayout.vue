<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import {
  DashboardOutlined,
  MessageOutlined,
  FileTextOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RobotOutlined,
  HistoryOutlined,
  ExperimentOutlined,
} from '@ant-design/icons-vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const appStore = useAppStore()
const route = useRoute()

const menuItems = [
  { key: '/', icon: DashboardOutlined, label: 'Dashboard' },
  { key: '/prompts', icon: MessageOutlined, label: 'Prompts' },
  { key: '/templates', icon: FileTextOutlined, label: 'Templates' },
  { key: '/generations', icon: HistoryOutlined, label: 'Generations' },
  { key: '/playground', icon: ExperimentOutlined, label: 'Playground' },
]

const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/': 'Dashboard',
    '/prompts': 'Prompts',
    '/templates': 'Templates',
    '/generations': 'Generation History',
    '/playground': 'Playground',
  }
  return map[route.path] || 'Dashboard'
})
</script>

<template>
  <a-layout class="app-layout">
    <a-layout-sider
      v-model:collapsed="appStore.sidebarCollapsed"
      :trigger="null"
      collapsible
      width="240"
      class="sidebar"
    >
      <div class="logo">
        <div class="logo-icon">
          <RobotOutlined />
        </div>
        <span v-if="!appStore.sidebarCollapsed" class="logo-text">Gemini</span>
      </div>
      <div class="menu-divider" />
      <a-menu :selected-keys="[route.path]" mode="inline" class="custom-menu" theme="light">
        <a-menu-item v-for="item in menuItems" :key="item.key">
          <RouterLink :to="item.key" class="menu-link">
            <component :is="item.icon" />
            <span>{{ item.label }}</span>
          </RouterLink>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-layout>
      <a-layout-header class="header">
        <div class="header-left">
          <a-button type="text" class="toggle-btn" @click="appStore.toggleSidebar">
            <MenuUnfoldOutlined v-if="appStore.sidebarCollapsed" />
            <MenuFoldOutlined v-else />
          </a-button>
          <div class="breadcrumb">
            <span class="breadcrumb-current">{{ pageTitle }}</span>
          </div>
        </div>
      </a-layout-header>

      <a-layout-content class="content-wrapper">
        <div class="content">
          <RouterView />
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.app-layout {
  height: 100vh;
  background: #f8fafc;
}

.sidebar {
  background: #ffffff !important;
  border-right: 1px solid #e2e8f0;
  z-index: 10;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar::-webkit-scrollbar {
  width: 4px;
}

.sidebar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 2px;
}

.sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.logo {
  height: 72px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 12px;
  transition: all 0.3s ease;
}

.logo-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #00d9a5;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0f172a;
  font-size: 18px;
  flex-shrink: 0;
}

.logo-text {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.menu-divider {
  height: 1px;
  background: #e2e8f0;
  margin: 0 20px 8px;
}

.custom-menu {
  border-right: none !important;
  background: transparent !important;
  padding: 8px 0;
}

.custom-menu :deep(.ant-menu-item) {
  transition: all 0.2s ease;
  color: #64748b !important;
}

.custom-menu :deep(.ant-menu-item-selected) {
  background: #f1f5f9 !important;
  color: #00d9a5 !important;
  font-weight: 600;
  box-shadow: none;
}

.custom-menu :deep(.ant-menu-item-selected::after) {
  display: none;
}

.custom-menu :deep(.ant-menu-item:hover) {
  background: #f1f5f9 !important;
  color: #00d9a5 !important;
}

.menu-link {
  display: flex;
  align-items: center;
  gap: 12px;
  color: inherit;
  text-decoration: none;
  width: 100%;
}

.header {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e2e8f0;
  height: 64px;
  position: sticky;
  top: 0;
  z-index: 9;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.toggle-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px !important;
  color: #64748b;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: #f1f5f9 !important;
  color: #0f172a;
}

.breadcrumb {
  display: flex;
  align-items: center;
}

.breadcrumb-current {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 18px;
  font-weight: 600;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.content-wrapper {
  padding: 0;
  background: transparent;
  overflow-y: auto;
}

.content {
  margin: 24px;
  padding: 24px;
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.04);
  border: 1px solid #e2e8f0;
}

/* Responsive */
@media (max-width: 768px) {
  .content {
    margin: 12px;
    padding: 16px;
    border-radius: 12px;
  }
}
</style>
