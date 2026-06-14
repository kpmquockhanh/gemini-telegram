<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { usePromptsStore } from '@/stores/prompts'
import { useTemplatesStore } from '@/stores/templates'
import { getProviders } from '@/api/client'
import { message } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import type { Prompt, PromptFormData, ProvidersResponse } from '@/types'

const promptsStore = usePromptsStore()
const templatesStore = useTemplatesStore()

const editModalVisible = ref(false)
const editForm = ref<PromptFormData>({ imagePrompt: '', videoPrompt: '', provider: '', modelName: '' })
const currentChatId = ref<number | null>(null)
const selectedTemplate = ref<number | null>(null)

const searchText = ref('')
const providers = ref<ProvidersResponse | null>(null)
const providersLoading = ref(false)

onMounted(() => {
  promptsStore.fetchPrompts()
  templatesStore.fetchTemplates()
  loadProviders()
})

async function loadProviders() {
  providersLoading.value = true
  try {
    providers.value = await getProviders()
  } catch (err) {
    console.error('Failed to load providers', err)
  } finally {
    providersLoading.value = false
  }
}

function getAvailableModels(providerName: string): string[] {
  if (!providers.value || !providers.value.providers[providerName]) return []
  return providers.value.providers[providerName].models
}

function handleSearch() {
  promptsStore.searchQuery = searchText.value
  promptsStore.fetchPrompts(1)
}

function handlePageChange(page: number, pageSize: number) {
  promptsStore.fetchPrompts(page, pageSize)
}

function openEditModal(record: Prompt) {
  currentChatId.value = record.chatId
  editForm.value = {
    imagePrompt: record.imagePrompt || '',
    videoPrompt: record.videoPrompt || '',
    provider: record.provider || providers.value?.default || '',
    modelName: record.modelName || '',
  }
  selectedTemplate.value = null
  editModalVisible.value = true
}

function applyTemplate() {
  if (selectedTemplate.value) {
    const template = templatesStore.items.find((t) => t.id === selectedTemplate.value)
    if (template) {
      editForm.value.imagePrompt = template.imagePrompt || editForm.value.imagePrompt
      editForm.value.videoPrompt = template.videoPrompt || editForm.value.videoPrompt
    }
  }
}

async function handleSave() {
  if (currentChatId.value === null) return
  const success = await promptsStore.savePrompt(currentChatId.value, editForm.value)
  if (success) {
    message.success('Prompt updated successfully')
    editModalVisible.value = false
  } else {
    message.error('Failed to update prompt')
  }
}

async function handleDelete(chatId: number) {
  const success = await promptsStore.removePrompt(chatId)
  if (success) {
    message.success('Prompt deleted successfully')
  } else {
    message.error('Failed to delete prompt')
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  message.success('Copied to clipboard')
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Prompts</h1>
        <p class="page-subtitle">Manage your chat prompts and templates</p>
      </div>
    </div>

    <div class="toolbar">
      <a-input-search
        v-model:value="searchText"
        placeholder="Search by chat ID or prompt content..."
        class="search-input"
        @search="handleSearch"
      >
        <template #prefix>
          <SearchOutlined />
        </template>
      </a-input-search>
      <div class="toolbar-actions">
        <a-button type="primary" @click="handleSearch">
          <SearchOutlined />
          Search
        </a-button>
        <a-button class="refresh-btn" @click="promptsStore.fetchPrompts()">
          <ReloadOutlined />
          Refresh
        </a-button>
      </div>
    </div>

    <a-table
      :data-source="promptsStore.items"
      :loading="promptsStore.loading"
      :pagination="{
        current: promptsStore.currentPage,
        pageSize: promptsStore.pageSize,
        total: promptsStore.total,
        showSizeChanger: true,
        pageSizeOptions: ['10', '20', '50', '100'],
      }"
      row-key="chatId"
      class="data-table"
      @change="handlePageChange"
    >
      <a-table-column title="Chat ID" data-index="chatId" key="chatId" width="100">
        <template #default="{ text }">
          <span class="chat-id">{{ text }}</span>
        </template>
      </a-table-column>
      <a-table-column title="Image Prompt" data-index="imagePrompt" key="imagePrompt" ellipsis>
        <template #default="{ text }">
          <div class="prompt-cell">
            <span v-if="text" class="prompt-text">{{ text }}</span>
            <span v-else class="empty-text">—</span>
            <span
              v-if="text"
              class="copy-btn"
              @click="copyToClipboard(text)"
            >
              <CopyOutlined />
            </span>
          </div>
        </template>
      </a-table-column>
      <a-table-column title="Video Prompt" data-index="videoPrompt" key="videoPrompt" ellipsis>
        <template #default="{ text }">
          <div class="prompt-cell">
            <span v-if="text" class="prompt-text">{{ text }}</span>
            <span v-else class="empty-text">—</span>
            <span
              v-if="text"
              class="copy-btn"
              @click="copyToClipboard(text)"
            >
              <CopyOutlined />
            </span>
          </div>
        </template>
      </a-table-column>
      <a-table-column title="Provider" data-index="provider" key="provider" width="120">
        <template #default="{ text, record }">
          <span v-if="text" class="provider-tag">{{ text }}</span>
          <span v-else class="empty-text">{{ providers?.default || '—' }}</span>
        </template>
      </a-table-column>
      <a-table-column title="Model" data-index="modelName" key="modelName" width="180">
        <template #default="{ text }">
          <span v-if="text" class="model-tag">{{ text }}</span>
          <span v-else class="empty-text">—</span>
        </template>
      </a-table-column>
      <a-table-column title="Updated" data-index="updatedAt" key="updatedAt" width="180">
        <template #default="{ text }">
          <span class="timestamp">{{ text }}</span>
        </template>
      </a-table-column>
      <a-table-column title="Actions" key="actions" width="120" align="center">
        <template #default="{ record }">
          <div class="action-btns">
            <a-button type="text" class="action-btn edit" size="small" @click="openEditModal(record)">
              <EditOutlined />
            </a-button>
            <a-popconfirm title="Are you sure you want to delete this prompt?" @confirm="handleDelete(record.chatId)">
              <a-button type="text" class="action-btn delete" size="small">
                <DeleteOutlined />
              </a-button>
            </a-popconfirm>
          </div>
        </template>
      </a-table-column>
    </a-table>

    <!-- Edit Modal -->
    <a-modal
      v-model:open="editModalVisible"
      title="Edit Prompt"
      ok-text="Save Changes"
      cancel-text="Cancel"
      class="custom-modal"
      @ok="handleSave"
    >
      <a-form layout="vertical">
        <a-form-item label="Chat ID">
          <a-input :value="currentChatId" disabled />
        </a-form-item>

        <a-form-item label="Apply Template">
          <a-select
            v-model:value="selectedTemplate"
            placeholder="Select a template"
            style="width: 100%"
            allow-clear
            @change="applyTemplate"
          >
            <a-select-option
              v-for="template in templatesStore.items"
              :key="template.id"
              :value="template.id"
            >
              {{ template.name }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="Provider">
          <a-select
            v-model:value="editForm.provider"
            placeholder="Select a provider"
            style="width: 100%"
            @change="editForm.modelName = ''"
          >
            <a-select-option
              v-for="(info, key) in providers?.providers"
              :key="key"
              :value="key"
            >
              {{ info.name }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="Model">
          <a-select
            v-model:value="editForm.modelName"
            placeholder="Select a model"
            style="width: 100%"
            :disabled="!editForm.provider"
          >
            <a-select-option
              v-for="model in getAvailableModels(editForm.provider)"
              :key="model"
              :value="model"
            >
              {{ model }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="Image Prompt">
          <a-textarea
            v-model:value="editForm.imagePrompt"
            :rows="4"
            :maxlength="2000"
            show-count
            placeholder="Enter default image prompt..."
          />
        </a-form-item>

        <a-form-item label="Video Prompt">
          <a-textarea
            v-model:value="editForm.videoPrompt"
            :rows="4"
            :maxlength="2000"
            show-count
            placeholder="Enter default video prompt..."
          />
        </a-form-item>
      </a-form>
    </a-modal>
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
  color: #0F172A;
  margin: 0;
  letter-spacing: -0.03em;
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
}

.refresh-btn {
  color: #475569;
}

.data-table {
  border-radius: 14px;
  overflow: hidden;
}

.chat-id {
  font-weight: 600;
  color: #00D9A5;
  font-family: 'SF Mono', monospace;
  font-size: 13px;
}

.prompt-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prompt-text {
  color: #0F172A;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-text {
  color: #94A3B8;
  font-style: italic;
}

.copy-btn {
  color: #00D9A5;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 6px;
  transition: all 0.15s;
  flex-shrink: 0;
  opacity: 0.6;
}

.copy-btn:hover {
  background: rgba(0, 217, 165, 0.1);
  opacity: 1;
}

.timestamp {
  color: #475569;
  font-size: 13px;
}

.action-btns {
  display: flex;
  gap: 4px;
  justify-content: center;
}

.action-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.15s;
}

.action-btn.edit {
  color: #00D9A5;
}

.action-btn.edit:hover {
  background: rgba(0, 217, 165, 0.1);
}

.action-btn.delete {
  color: #EF4444;
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.1);
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

.provider-tag {
  font-weight: 600;
  color: #00D9A5;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
  background: rgba(0, 217, 165, 0.1);
  padding: 2px 8px;
  border-radius: 6px;
}

.model-tag {
  font-weight: 500;
  color: #60A5FA;
  font-family: 'SF Mono', monospace;
  font-size: 12px;
  background: rgba(96, 165, 250, 0.1);
  padding: 2px 8px;
  border-radius: 6px;
}
</style>
