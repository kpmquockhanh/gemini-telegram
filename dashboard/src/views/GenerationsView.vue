<script setup lang="ts">
import { ref, computed } from 'vue'
import { useLogsStore } from '@/stores/logs'
import {
  SearchOutlined,
  ReloadOutlined,
  DeleteOutlined,
  HistoryOutlined,
} from '@ant-design/icons-vue'
// import type { GenerationHistory } from '@/types'
import { message, Modal } from 'ant-design-vue'

const logsStore = useLogsStore()

const searchText = ref('')
const statusFilter = ref<'all' | 'success' | 'error'>('all')
const expandedRows = ref<Set<string>>(new Set())

const filteredHistory = computed(() => {
  let history = logsStore.generationHistory
  if (statusFilter.value !== 'all') {
    history = history.filter((h) => h.status === statusFilter.value)
  }
  if (searchText.value.trim()) {
    const q = searchText.value.toLowerCase()
    history = history.filter(
      (h) =>
        h.operation.toLowerCase().includes(q) ||
        JSON.stringify(h.params).toLowerCase().includes(q),
    )
  }
  return history
})

function formatTime(ts: string) {
  return new Date(ts).toLocaleString()
}

function toggleExpand(id: string) {
  const next = new Set(expandedRows.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  expandedRows.value = next
}

function isExpanded(id: string) {
  return expandedRows.value.has(id)
}

function handleClear() {
  Modal.confirm({
    title: 'Clear generation history?',
    content: 'This will permanently remove all generation history. This action cannot be undone.',
    okText: 'Clear',
    okType: 'danger',
    onOk: () => {
      logsStore.clearGenerationHistory()
      message.success('Generation history cleared')
    },
  })
}

function prettyJson(obj: unknown): string {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">
          <HistoryOutlined />
          Generation History
        </h1>
        <p class="page-subtitle">Local record of requests, parameters, and results</p>
      </div>
    </div>

    <div class="toolbar">
      <a-input-search
        v-model:value="searchText"
        placeholder="Search operations..."
        class="search-input"
        allow-clear
      >
        <template #prefix>
          <SearchOutlined />
        </template>
      </a-input-search>

      <div class="toolbar-actions">
        <a-select v-model:value="statusFilter" style="width: 120px">
          <a-select-option value="all">All Status</a-select-option>
          <a-select-option value="success">Success</a-select-option>
          <a-select-option value="error">Error</a-select-option>
        </a-select>

        <a-button class="refresh-btn" @click="logsStore.generationHistory">
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
      :data-source="filteredHistory"
      row-key="id"
      class="data-table"
      :pagination="{ pageSize: 20, showSizeChanger: true, pageSizeOptions: ['10', '20', '50', '100'] }"
    >
      <a-table-column title="Time" key="timestamp" width="180">
        <template #default="{ record }">
          <span class="timestamp">{{ formatTime(record.timestamp) }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Operation" key="operation" width="160">
        <template #default="{ record }">
          <span class="operation-tag">{{ record.operation }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Status" key="status" width="110">
        <template #default="{ record }">
          <a-tag :color="record.status === 'success' ? 'green' : 'red'" class="status-tag">
            {{ record.status.toUpperCase() }}
          </a-tag>
        </template>
      </a-table-column>

      <a-table-column title="Params" key="params">
        <template #default="{ record }">
          <div class="code-preview" @click="toggleExpand(record.id + '-params')">
            <pre><code>{{ isExpanded(record.id + '-params') ? prettyJson(record.params) : prettyJson(record.params).slice(0, 120) }}{{ prettyJson(record.params).length > 120 && !isExpanded(record.id + '-params') ? '...' : '' }}</code></pre>
          </div>
        </template>
      </a-table-column>

      <a-table-column title="Result" key="result">
        <template #default="{ record }">
          <div class="code-preview" @click="toggleExpand(record.id + '-result')">
            <pre><code>{{ isExpanded(record.id + '-result') ? prettyJson(record.result) : prettyJson(record.result).slice(0, 120) }}{{ prettyJson(record.result).length > 120 && !isExpanded(record.id + '-result') ? '...' : '' }}</code></pre>
          </div>
        </template>
      </a-table-column>

      <a-table-column title="Duration" key="durationMs" width="100" align="right">
        <template #default="{ record }">
          <span v-if="record.durationMs" class="duration">{{ record.durationMs }}ms</span>
          <span v-else class="empty-text">—</span>
        </template>
      </a-table-column>
    </a-table>

    <a-empty v-if="filteredHistory.length === 0" description="No generation history found" class="empty-state" />
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

.operation-tag {
  font-weight: 600;
  color: #00d9a5;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
  background: rgba(0, 217, 165, 0.1);
  padding: 2px 8px;
  border-radius: 6px;
}

.status-tag {
  font-weight: 600;
  font-size: 11px;
  letter-spacing: 0.05em;
}

.code-preview {
  max-width: 320px;
  cursor: pointer;
}

.code-preview pre {
  margin: 0;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px 10px;
  font-size: 12px;
  color: #475569;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 120px;
  overflow: auto;
}

.duration {
  font-weight: 600;
  color: #0f172a;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
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
