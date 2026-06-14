<script setup lang="ts">
import { onMounted } from 'vue'
import { usePromptsStore } from '@/stores/prompts'
import { getStats } from '@/api/client'
import { ref } from 'vue'
import type { Stats } from '@/types'
import {
  RobotOutlined,
  PictureOutlined,
  VideoCameraOutlined,
  MessageOutlined,
  ArrowUpOutlined,
} from '@ant-design/icons-vue'

const promptsStore = usePromptsStore()
const stats = ref<Stats>({
  totalChats: 0,
  imageCount: 0,
  videoCount: 0,
  totalPrompts: 0,
})
const statsLoading = ref(false)

async function loadStats() {
  statsLoading.value = true
  try {
    stats.value = await getStats()
  } catch (err) {
    console.error('Failed to load stats', err)
  } finally {
    statsLoading.value = false
  }
}

onMounted(() => {
  promptsStore.fetchPrompts(1, 10)
  loadStats()
})

const statCards = [
  {
    title: 'Total Chats',
    key: 'totalChats' as keyof Stats,
    icon: RobotOutlined,
    accent: '#00D9A5',
    glow: 'rgba(0, 217, 165, 0.15)',
  },
  {
    title: 'Image Prompts',
    key: 'imageCount' as keyof Stats,
    icon: PictureOutlined,
    accent: '#60A5FA',
    glow: 'rgba(96, 165, 250, 0.15)',
  },
  {
    title: 'Video Prompts',
    key: 'videoCount' as keyof Stats,
    icon: VideoCameraOutlined,
    accent: '#F59E0B',
    glow: 'rgba(245, 158, 11, 0.15)',
  },
  {
    title: 'Total Prompts',
    key: 'totalPrompts' as keyof Stats,
    icon: MessageOutlined,
    accent: '#A78BFA',
    glow: 'rgba(167, 139, 250, 0.15)',
  },
]
</script>

<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Dashboard</h1>
      <p class="page-subtitle">AI generation engine overview</p>
    </div>

    <a-row :gutter="[16, 16]" style="margin-top: 24px">
      <a-col
        v-for="(card, index) in statCards"
        :key="card.key"
        :xs="24"
        :sm="12"
        :lg="6"
      >
        <div class="artifact-card" :style="{ '--accent': card.accent, '--glow': card.glow }">
          <div class="artifact-border"></div>
          <div class="artifact-content">
            <div class="artifact-meta">
              <div class="artifact-icon" :style="{ background: card.accent + '15', color: card.accent }">
                <component :is="card.icon" />
              </div>
              <div class="artifact-status">
                <ArrowUpOutlined />
                <span>Active</span>
              </div>
            </div>
            <div class="artifact-data">
              <div class="artifact-value">{{ stats[card.key] || 0 }}</div>
              <div class="artifact-label">{{ card.title }}</div>
            </div>
          </div>
        </div>
      </a-col>
    </a-row>

    <div class="section-header">
      <div class="section-meta">
        <h3 class="section-title">Recent Prompts</h3>
        <p class="section-desc">Latest prompt updates across chats</p>
      </div>
      <a-button type="primary" @click="$router.push('/prompts')">
        View All
      </a-button>
    </div>

    <a-table
      :columns="[
        { title: 'Chat ID', dataIndex: 'chatId', key: 'chatId', width: 100 },
        { title: 'Image Prompt', dataIndex: 'imagePrompt', key: 'imagePrompt', ellipsis: true },
        { title: 'Video Prompt', dataIndex: 'videoPrompt', key: 'videoPrompt', ellipsis: true },
        { title: 'Provider', dataIndex: 'provider', key: 'provider', width: 100 },
        { title: 'Model', dataIndex: 'modelName', key: 'modelName', width: 160 },
        { title: 'Updated', dataIndex: 'updatedAt', key: 'updatedAt', width: 180 },
      ]"
      :data-source="promptsStore.items"
      :loading="promptsStore.loading"
      :pagination="false"
      row-key="chatId"
      size="small"
      class="recent-table"
    />
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: 8px;
}
.page-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 28px;
  font-weight: 700;
  color: #F8FAFC;
  margin: 0;
  letter-spacing: -0.03em;
}
.page-subtitle {
  color: #8B95A8;
  margin: 6px 0 0;
  font-size: 15px;
}

/* ── AI Artifact Cards ── */
.artifact-card {
  position: relative;
  border-radius: 16px;
  background: #111827;
  overflow: hidden;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  cursor: default;
}

.artifact-card:hover {
  transform: translateY(-2px);
}

.artifact-border {
  position: absolute;
  inset: 0;
  border-radius: 16px;
  padding: 1px;
  background: linear-gradient(
    135deg,
    var(--accent) 0%,
    transparent 40%,
    transparent 60%,
    var(--accent) 100%
  );
  opacity: 0.25;
  transition: opacity 0.3s ease;
  pointer-events: none;
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
}

.artifact-card:hover .artifact-border {
  opacity: 0.6;
}

.artifact-content {
  padding: 24px;
  position: relative;
  z-index: 1;
}

.artifact-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.artifact-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.artifact-card:hover .artifact-icon {
  transform: scale(1.05);
  box-shadow: 0 0 20px var(--glow);
}

.artifact-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--accent);
  background: var(--glow);
  padding: 4px 10px;
  border-radius: 20px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.artifact-data {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.artifact-value {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 32px;
  font-weight: 700;
  color: #F8FAFC;
  line-height: 1;
  letter-spacing: -0.02em;
}

.artifact-label {
  font-size: 13px;
  font-weight: 500;
  color: #8B95A8;
}

/* ── Section ── */
.section-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-top: 40px;
  margin-bottom: 20px;
  gap: 16px;
}

.section-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 20px;
  font-weight: 600;
  color: #F8FAFC;
  margin: 0;
  letter-spacing: -0.02em;
}

.section-desc {
  color: #8B95A8;
  font-size: 13px;
  margin: 0;
}

.recent-table {
  border-radius: 14px;
  overflow: hidden;
}

/* Responsive */
@media (max-width: 768px) {
  .page-title {
    font-size: 24px;
  }
  .artifact-content {
    padding: 20px;
  }
  .artifact-value {
    font-size: 28px;
  }
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
