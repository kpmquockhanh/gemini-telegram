<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { usePromptsStore } from '@/stores/prompts'
import { useTemplatesStore } from '@/stores/templates'
import { getProviders } from '@/api/client'
import { message } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import type { Prompt, PromptFormData, ProvidersResponse, ModelParamDef } from '@/types'

const promptsStore = usePromptsStore()
const templatesStore = useTemplatesStore()

const editModalVisible = ref(false)
const editForm = ref<PromptFormData>({ templateId: null, provider: '', modelName: '', params: {} })
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
  return providers.value.providers[providerName].models.map((m) => m.name)
}

function getModelParams(providerName: string, modelName: string): ModelParamDef[] {
  if (!providers.value || !providerName || !modelName) return []
  const provider = providers.value.providers[providerName]
  if (!provider) return []
  const model = provider.models.find((m) => m.name === modelName)
  return model?.params || []
}

function getParamDefault(def: ModelParamDef): unknown {
  return def.default
}

function onModelChange() {
  const defaults: Record<string, unknown> = {}
  for (const param of getModelParams(editForm.value.provider, editForm.value.modelName)) {
    defaults[param.name] = param.default
  }
  editForm.value.params = defaults
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
  let params: Record<string, unknown> = {}
  if (record.params) {
    try {
      params = JSON.parse(record.params)
    } catch {
      params = {}
    }
  }
  editForm.value = {
    templateId: record.templateId || null,
    provider: record.provider || providers.value?.default || '',
    modelName: record.modelName || '',
    params,
  }
  selectedTemplate.value = record.templateId || null
  editModalVisible.value = true
}

async function handleSave() {
  if (currentChatId.value === null) return
  
  // Ensure templateId is set from the selector
  if (selectedTemplate.value !== null) {
    editForm.value.templateId = selectedTemplate.value
  }
  
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

function getTemplateName(templateId: number | null): string {
  if (!templateId) return '—'
  const template = templatesStore.items.find((t) => t.id === templateId)
  return template ? template.name : `#${templateId}`
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Prompts</h1>
        <p class="page-subtitle">Manage your chat templates and providers</p>
      </div>
    </div>

    <div class="toolbar">
      <a-input-search
        v-model:value="searchText"
        placeholder="Search by chat ID or template name..."
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
      <a-table-column title="Template" data-index="templateId" key="templateId" width="200">
        <template #default="{ text, record }">
          <span v-if="record.templateName" class="template-name">{{ record.templateName }}</span>
          <span v-else class="empty-text">—</span>
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

        <a-form-item label="Template" required>
          <a-select
            v-model:value="selectedTemplate"
            placeholder="Select a template"
            style="width: 100%"
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
            @change="editForm.modelName = ''; editForm.params = {}"
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
            @change="onModelChange"
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

        <a-form-item
          v-for="param in getModelParams(editForm.provider, editForm.modelName)"
          :key="param.name"
          :label="param.label"
        >
          <a-slider
            v-if="param.type === 'slider'"
            :min="param.min ?? 0"
            :max="param.max ?? 1"
            :step="param.step ?? 0.1"
            :value="(editForm.params[param.name] as number) ?? (param.default as number)"
            @change="(v: number) => editForm.params[param.name] = v"
          />
          <a-input-number
            v-else-if="param.type === 'number'"
            :min="param.min ?? undefined"
            :max="param.max ?? undefined"
            :step="param.step ?? 1"
            :value="(editForm.params[param.name] as number) ?? (param.default as number)"
            style="width: 100%"
            @change="(v: number | null) => { if (v !== null) editForm.params[param.name] = v }"
          />
          <a-select
            v-else-if="param.type === 'select'"
            :value="(editForm.params[param.name] as string) ?? (param.default as string)"
            style="width: 100%"
            @change="(v: string) => editForm.params[param.name] = v"
          >
            <a-select-option
              v-for="opt in param.options"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <!-- Template Preview -->
        <div v-if="selectedTemplate" class="template-preview">
          <div class="preview-header">
            <span class="preview-label">Template Preview</span>
          </div>
          <div v-if="templatesStore.items.find(t => t.id === selectedTemplate)?.imagePrompt" class="preview-section">
            <span class="preview-type">Image Prompt</span>
            <p class="preview-text">
              {{ templatesStore.items.find(t => t.id === selectedTemplate)?.imagePrompt }}
            </p>
          </div>
          <div v-if="templatesStore.items.find(t => t.id === selectedTemplate)?.videoPrompt" class="preview-section">
            <span class="preview-type">Video Prompt</span>
            <p class="preview-text">
              {{ templatesStore.items.find(t => t.id === selectedTemplate)?.videoPrompt }}
            </p>
          </div>
        </div>
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

.template-name {
  font-weight: 600;
  color: #0F172A;
}

.empty-text {
  color: #94A3B8;
  font-style: italic;
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

/* Template Preview */
.template-preview {
  margin-top: 16px;
  padding: 16px;
  background: #F8FAFC;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
}

.preview-header {
  margin-bottom: 12px;
}

.preview-label {
  font-weight: 600;
  font-size: 13px;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.preview-section {
  margin-bottom: 12px;
}

.preview-section:last-child {
  margin-bottom: 0;
}

.preview-type {
  font-weight: 600;
  font-size: 12px;
  color: #00D9A5;
  margin-bottom: 4px;
  display: block;
}

.preview-text {
  font-size: 13px;
  color: #64748B;
  margin: 0;
  line-height: 1.5;
  word-break: break-word;
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
