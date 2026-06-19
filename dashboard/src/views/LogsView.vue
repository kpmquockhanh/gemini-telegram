<script setup lang="ts">
import { ref, computed } from 'vue'
import { useLogsStore } from '@/stores/logs'
import {
  SearchOutlined,
  ReloadOutlined,
  DeleteOutlined,
  BugOutlined,
} from '@ant-design/icons-vue'
import type { TraceLog } from '@/types'
import { message, Modal } from 'ant-design-vue'

const logsStore = useLogsStore()

const searchText = ref('')
const levelFilter = ref<'all' | 'info' | 'warn' | 'error'>('all')

const filteredLogs = computed(() => {
  let logs = logsStore.traceLogs
  if (levelFilter.value !== 'all') {
    logs = logs.filter((l) => l.level === levelFilter.value)
  }
  if (searchText.value.trim()) {
    const q = searchText.value.toLowerCase()
    logs = logs.filter(
      (l) =>
        l.message.toLowerCase().includes(q) ||
        JSON.stringify(l.context).toLowerCase().includes(q),
    )
  }
  return logs
})

function formatTime(ts: string) {
  return new Date(ts).toLocaleString()
}

function getLevelColor(level: TraceLog['level']) {
  switch (level) {
    case 'error':
      return 'red'
    case 'warn':
      return 'orange'
    default:
      return 'green'
  }
}

function handleClear() {
  Modal.confirm({
    title: 'Clear trace logs?',
    content: 'This will permanently remove all trace logs. This action cannot be undone.',
    okText: 'Clear',
    okType: 'danger',
    onOk: () => {
      logsStore.clearTraceLogs()
      message.success('Trace logs cleared')
    },
  })
}

function contextSummary(ctx?: Record<string, unknown>): string {
  if (!ctx) return '—'
  return JSON.stringify(ctx)
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">
          <BugOutlined />
          Trace Logs
        </h1>
        <p class="page-subtitle">Local trace logs captured from application activity</p>
      </div>
    </div>

    <div class="toolbar">
      <a-input-search
        v-model:value="searchText"
        placeholder="Search logs..."
        class="search-input"
        allow-clear
      >
        <template #prefix>
          <SearchOutlined />
        </template>
      </a-input-search>

      <div class="toolbar-actions">
        <a-select v-model:value="levelFilter" style="width: 120px">
          <a-select-option value="all">All Levels</a-select-option>
          <a-select-option value="info">Info</a-select-option>
          <a-select-option value="warn">Warn</a-select-option>
          <a-select-option value="error">Error</a-select-option>
        </a-select>

        <a-button class="refresh-btn" @click="logsStore.traceLogs">
          <ReloadOutlined />
          Refresh
        </a-button>

        <a-button danger @click="handleClear">
          <DeleteOutlined />
          Clear
        </a-button>
      </div>
    </div>

    <a-table
      :data-source="filteredLogs"
      row-key="id"
      class="data-table"
      :pagination="{ pageSize: 20, showSizeChanger: true, pageSizeOptions: ['10', '20', '50', '100'] }"
    >
      <a-table-column title="Time" key="timestamp" width="180">
        <template #default="{ record }">
          <span class="timestamp">{{ formatTime(record.timestamp) }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Level" key="level" width="100">
        <template #default="{ record }">
          <a-tag :color="getLevelColor(record.level)" class="level-tag">
            {{ record.level.toUpperCase() }}
          </a-tag>
        </template>
      </a-table-column>

      <a-table-column title="Message" key="message">
        <template #default="{ record }">
          <span class="message-text">{{ record.message }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Context" key="context" width="300">
        <template #default="{ record }">
          <div class="context-cell">
            <a-typography-text
              v-if="record.context && Object.keys(record.context).length"
              code
              copyable
              class="context-code"
            >
              {{ contextSummary(record.context) }}
            </a-typography-text>
            <span v-else class="empty-text">—</span>
          </div>
        </template>
      </a-table-column>
    </a-table>

    <a-empty v-if="filteredLogs.length === 0" description="No trace logs found" class="empty-state" />
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: 24px;
}
.page-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  letter-spacing: -0.03em;
  display: flex;
  align-items: center;
  gap: 12px;
}
.page-subtitle {
  color: #475569;
  margin: 6px 0 0;
  font-size: 15px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.search-input {
  width: 360px;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.refresh-btn {
  color: #475569;
}

.data-table {
  border-radius: 14px;
  overflow: hidden;
}

.timestamp {
  color: #475569;
  font-size: 13px;
  font-family: 'SF Mono', monospace;
}

.level-tag {
  font-weight: 600;
  font-size: 11px;
  letter-spacing: 0.05em;
}

.message-text {
  color: #0f172a;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.context-cell {
  max-width: 300px;
}

.context-code {
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.empty-text {
  color: #94a3b8;
  font-style: italic;
}

.empty-state {
  margin-top: 48px;
}

/* Responsive */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  .search-input {
    width: 100%;
  }
  .toolbar-actions {
    justify-content: flex-end;
  }
}
</style>
