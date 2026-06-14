<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import {
  DashboardOutlined,
  MessageOutlined,
  FileTextOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RobotOutlined,
} from '@ant-design/icons-vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const appStore = useAppStore()
const route = useRoute()

const menuItems = [
  { key: '/', icon: DashboardOutlined, label: 'Dashboard' },
  { key: '/prompts', icon: MessageOutlined, label: 'Prompts' },
  { key: '/templates', icon: FileTextOutlined, label: 'Templates' },
]

const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/': 'Dashboard',
    '/prompts': 'Prompts',
    '/templates': 'Templates',
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
      <a-menu
        :selected-keys="[route.path]"
        mode="inline"
        class="custom-menu"
        theme="light"
      >
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
          <a-button
            type="text"
            class="toggle-btn"
            @click="appStore.toggleSidebar"
          >
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
  min-height: 100vh;
  background: #F8FAFC;
}

.sidebar {
  background: #FFFFFF !important;
  border-right: 1px solid #E2E8F0;
  z-index: 10;
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
  background: #00D9A5;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0F172A;
  font-size: 18px;
  flex-shrink: 0;
}

.logo-text {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 20px;
  font-weight: 700;
  color: #0F172A;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.menu-divider {
  height: 1px;
  background: #E2E8F0;
  margin: 0 20px 8px;
}

.custom-menu {
  border-right: none !important;
  background: transparent !important;
  padding: 8px 0;
}

.custom-menu :deep(.ant-menu-item) {
  border-radius: 10px !important;
  margin: 4px 8px !important;
  padding: 0 12px !important;
  height: 44px;
  line-height: 44px;
  transition: all 0.2s ease;
  color: #64748B !important;
}

.custom-menu :deep(.ant-menu-item-selected) {
  background: #F1F5F9 !important;
  color: #00D9A5 !important;
  font-weight: 600;
  box-shadow: none;
}

.custom-menu :deep(.ant-menu-item-selected::after) {
  display: none;
}

.custom-menu :deep(.ant-menu-item:hover) {
  background: #F1F5F9 !important;
  color: #00D9A5 !important;
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
  border-bottom: 1px solid #E2E8F0;
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
  color: #64748B;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: #F1F5F9 !important;
  color: #0F172A;
}

.breadcrumb {
  display: flex;
  align-items: center;
}

.breadcrumb-current {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 18px;
  font-weight: 600;
  color: #0F172A;
  letter-spacing: -0.02em;
}

.content-wrapper {
  padding: 0;
  background: transparent;
}

.content {
  margin: 24px;
  padding: 24px;
  background: #FFFFFF;
  border-radius: 16px;
  min-height: calc(100vh - 112px);
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.04);
  border: 1px solid #E2E8F0;
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
