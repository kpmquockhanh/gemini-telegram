<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { listGenerationHistory, clearGenerationHistory } from '@/api/client'
import {
  SearchOutlined,
  ReloadOutlined,
  DeleteOutlined,
  HistoryOutlined,
} from '@ant-design/icons-vue'
import type { GenerationHistory } from '@/types'
import { message, Modal } from 'ant-design-vue'

const history = ref<GenerationHistory[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const loading = ref(false)

const searchText = ref('')
const statusFilter = ref<'all' | 'success' | 'error'>('all')
const expandedRows = ref<Set<string>>(new Set())

const filteredHistory = computed(() => history.value ?? [])

async function fetchHistory() {
  loading.value = true
  try {
    const res = await listGenerationHistory(
      page.value,
      limit.value,
      statusFilter.value,
      searchText.value,
    )
    history.value = res.items
    total.value = res.total
  } catch (err) {
    message.error('Failed to load generation history')
    console.error(err)
  } finally {
    loading.value = false
  }
}

function onPageChange(newPage: number, newPageSize: number) {
  page.value = newPage
  limit.value = newPageSize
  fetchHistory()
}

function onShowSizeChange(newPage: number, newPageSize: number) {
  page.value = newPage
  limit.value = newPageSize
  fetchHistory()
}

watch([statusFilter], () => {
  page.value = 1
  fetchHistory()
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
    onOk: async () => {
      try {
        await clearGenerationHistory()
        message.success('Generation history cleared')
        page.value = 1
        await fetchHistory()
      } catch (err) {
        message.error('Failed to clear generation history')
        console.error(err)
      }
    },
  })
}

function prettyJson(str: string): string {
  if (!str) return ''
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

fetchHistory()
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">
          <HistoryOutlined />
          Generation History
        </h1>
        <p class="page-subtitle">Database record of image/video generation requests</p>
      </div>
    </div>

    <div class="toolbar">
      <a-input-search
        v-model:value="searchText"
        placeholder="Search prompts, providers, models..."
        class="search-input"
        allow-clear
        @search="fetchHistory"
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

        <a-button class="refresh-btn" @click="fetchHistory">
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
      :loading="loading"
      :pagination="false"
    >
      <a-table-column title="Time" key="createdAt" width="180">
        <template #default="{ record }">
          <span class="timestamp">{{ formatTime(record.createdAt) }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Chat" key="chatId" width="100">
        <template #default="{ record }">
          <span class="chat-id">{{ record.chatId }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Type" key="jobType" width="100">
        <template #default="{ record }">
          <span class="type-tag">{{ record.jobType }}</span>
        </template>
      </a-table-column>

      <a-table-column title="Prompt" key="prompt" width="200">
        <template #default="{ record }">
          <div class="prompt-cell">{{ record.prompt }}</div>
        </template>
      </a-table-column>

      <a-table-column title="Provider / Model" key="provider" width="180">
        <template #default="{ record }">
          <div class="provider-cell">
            <div class="provider-name">{{ record.provider || '—' }}</div>
            <div class="model-name">{{ record.modelName || '—' }}</div>
          </div>
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

    <div class="pagination-bar">
      <a-pagination
        v-model:current="page"
        v-model:page-size="limit"
        :total="total"
        :page-size-options="['10', '20', '50', '100']"
        show-size-changer
        :show-total="(total: number, range: [number, number]) => `Total ${total} items`"
        @change="onPageChange"
        @show-size-change="onShowSizeChange"
      />
    </div>

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

.chat-id {
  font-weight: 600;
  color: #0f172a;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
}

.type-tag {
  font-weight: 600;
  color: #00d9a5;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
  background: rgba(0, 217, 165, 0.1);
  padding: 2px 8px;
  border-radius: 6px;
  text-transform: uppercase;
}

.prompt-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #0f172a;
  font-size: 13px;
}

.provider-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.provider-name {
  font-weight: 600;
  color: #0f172a;
  font-size: 12px;
}

.model-name {
  color: #475569;
  font-size: 11px;
  font-family: 'SF Mono', monospace;
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

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
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
