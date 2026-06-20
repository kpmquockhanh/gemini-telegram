<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { getProviders, generate } from '@/api/client'
import type { ProvidersResponse, ProviderInfo, ModelInfo, ModelParamDef, GenerateResponse } from '@/types'
import {
  ThunderboltOutlined,
  ExperimentOutlined,
  SendOutlined,
  ClearOutlined,
  PictureOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const providersData = ref<ProvidersResponse | null>(null)
const providersLoading = ref(false)

const selectedProvider = ref('')
const selectedModel = ref('')
const genType = ref<'image' | 'video'>('image')
const prompt = ref('')
const params = ref<Record<string, unknown>>({})
const generating = ref(false)
const result = ref<GenerateResponse | null>(null)
const error = ref<string | null>(null)

const referenceImage = ref<string | null>(null)
const referenceMimeType = ref('')
const referenceFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

const referencePreview = computed(() => {
  if (!referenceImage.value) return null
  return `data:${referenceMimeType.value || 'image/jpeg'};base64,${referenceImage.value}`
})

const providerOptions = computed(() => {
  if (!providersData.value) return []
  return Object.keys(providersData.value.providers).map((key) => ({
    value: key,
    label: key,
  }))
})

const currentProvider = computed<ProviderInfo | null>(() => {
  if (!providersData.value || !selectedProvider.value) return null
  return providersData.value.providers[selectedProvider.value] || null
})

const modelOptions = computed(() => {
  if (!currentProvider.value) return []
  return currentProvider.value.models.map((m: ModelInfo) => ({
    value: m.name,
    label: m.name,
  }))
})

const currentModel = computed<ModelInfo | null>(() => {
  if (!currentProvider.value || !selectedModel.value) return null
  return currentProvider.value.models.find((m) => m.name === selectedModel.value) || null
})

const paramDefs = computed<ModelParamDef[]>(() => {
  return currentModel.value?.params || []
})

function defaultValue(def: ModelParamDef): unknown {
  if (def.default !== undefined && def.default !== null) return def.default
  if (def.type === 'slider' || def.type === 'number') return def.min ?? 0
  if (def.type === 'select' && def.options && def.options.length > 0) return def.options[0]!.value
  return ''
}

function initParams() {
  const p: Record<string, unknown> = {}
  for (const def of paramDefs.value) {
    p[def.name] = params.value[def.name] ?? defaultValue(def)
  }
  params.value = p
}

watch(selectedProvider, () => {
  if (currentProvider.value) {
    selectedModel.value = currentProvider.value.models[0]?.name || ''
  } else {
    selectedModel.value = ''
  }
})

watch(selectedModel, () => {
  initParams()
})

watch(currentProvider, () => {
  initParams()
})

async function loadProviders() {
  providersLoading.value = true
  try {
    providersData.value = await getProviders()
    if (providersData.value.default) {
      selectedProvider.value = providersData.value.default
    }
  } catch (err) {
    console.error('Failed to load providers', err)
  } finally {
    providersLoading.value = false
  }
}

function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = () => {
    const base64 = (reader.result as string).split(',')[1] ?? null
    referenceImage.value = base64
    referenceMimeType.value = file.type
    referenceFileName.value = file.name
  }
  reader.readAsDataURL(file)
}

function handleRemoveReference() {
  referenceImage.value = null
  referenceMimeType.value = ''
  referenceFileName.value = ''
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

async function handleGenerate() {
  if (!prompt.value.trim()) {
    message.warning('Please enter a prompt')
    return
  }
  generating.value = true
  result.value = null
  error.value = null

  try {
    const res = await generate({
      type: genType.value,
      prompt: prompt.value.trim(),
      provider: selectedProvider.value,
      modelName: selectedModel.value,
      params: Object.keys(params.value).length > 0 ? params.value : undefined,
      referenceImage: referenceImage.value || undefined,
      referenceMimeType: referenceMimeType.value || undefined,
    })
    result.value = res
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || 'Generation failed'
    error.value = msg
  } finally {
    generating.value = false
  }
}

function handleClear() {
  result.value = null
  error.value = null
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

const paramInputStyle = computed(() => {
  const n = paramDefs.value.length
  const width = n <= 2 ? '50%' : n <= 3 ? '33.33%' : '25%'
  return { width }
})

onMounted(() => {
  loadProviders()
})
</script>

<template>
  <div>
    <div class="page-header">
      <div class="header-icon">
        <ExperimentOutlined />
      </div>
      <div>
        <h1 class="page-title">Playground</h1>
        <p class="page-subtitle">Test AI generation with custom prompts, providers, and models</p>
      </div>
    </div>

    <a-row :gutter="[24, 24]" style="margin-top: 24px">
      <!-- Input Panel -->
      <a-col :xs="24" :lg="12">
        <div class="panel">
          <div class="panel-header">
            <ThunderboltOutlined />
            <span>Generation Config</span>
          </div>

          <a-form layout="vertical" class="gen-form">
            <a-form-item label="Type">
              <a-radio-group v-model:value="genType" button-style="solid" size="small">
                <a-radio-button value="image">Image</a-radio-button>
                <a-radio-button value="video">Video</a-radio-button>
              </a-radio-group>
            </a-form-item>

            <a-form-item label="Prompt">
              <a-textarea
                v-model:value="prompt"
                :rows="3"
                placeholder="Describe what you want to generate..."
                :disabled="generating"
              />
            </a-form-item>

            <a-form-item label="Reference Image (optional)">
              <div v-if="referencePreview" class="ref-preview">
                <img :src="referencePreview" class="ref-thumb" />
                <div class="ref-info">
                  <span class="ref-name">{{ referenceFileName }}</span>
                  <a-button
                    type="text"
                    danger
                    size="small"
                    :disabled="generating"
                    @click="handleRemoveReference"
                  >
                    <template #icon><DeleteOutlined /></template>
                  </a-button>
                </div>
              </div>
              <div v-else class="ref-upload" @click="fileInputRef?.click()">
                <input
                  ref="fileInputRef"
                  type="file"
                  accept="image/*"
                  style="display: none"
                  :disabled="generating"
                  @change="handleFileSelect"
                />
                <PictureOutlined :style="{ fontSize: '24px', color: '#94A3B8' }" />
                <span>Click to upload a reference image</span>
              </div>
            </a-form-item>

            <a-form-item label="Provider">
              <a-select
                v-model:value="selectedProvider"
                :options="providerOptions"
                :loading="providersLoading"
                :disabled="generating"
                placeholder="Select provider"
              />
            </a-form-item>

            <a-form-item label="Model">
              <a-select
                v-model:value="selectedModel"
                :options="modelOptions"
                :disabled="generating || !currentProvider"
                placeholder="Select model"
              />
            </a-form-item>

            <a-form-item v-if="paramDefs.length > 0" label="Parameters">
              <div class="param-grid">
                <div
                  v-for="def in paramDefs"
                  :key="def.name"
                  :style="paramInputStyle"
                  class="param-item"
                >
                  <label class="param-label">{{ def.label }}</label>

                  <template v-if="def.type === 'select'">
                    <a-select
                      v-model:value="params[def.name]"
                      :options="(def.options || []).map(o => ({ value: o.value, label: o.label }))"
                      :disabled="generating"
                      size="small"
                    />
                  </template>

                  <template v-else-if="def.type === 'slider'">
                    <a-slider
                      v-model:value="params[def.name]"
                      :min="def.min ?? 0"
                      :max="def.max ?? 1"
                      :step="def.step ?? 0.01"
                      :disabled="generating"
                    />
                  </template>

                  <template v-else>
                    <a-input-number
                      v-model:value="params[def.name]"
                      :min="def.min"
                      :max="def.max"
                      :step="def.step ?? 1"
                      :disabled="generating"
                      size="small"
                      style="width: 100%"
                    />
                  </template>
                </div>
              </div>
            </a-form-item>

            <a-form-item>
              <a-space>
                <a-button
                  type="primary"
                  :loading="generating"
                  :disabled="!prompt.trim()"
                  @click="handleGenerate"
                >
                  <template #icon><SendOutlined /></template>
                  Generate
                </a-button>
                <a-button :disabled="generating" @click="handleClear">
                  <template #icon><ClearOutlined /></template>
                  Clear
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>
        </div>
      </a-col>

      <!-- Result Panel -->
      <a-col :xs="24" :lg="12">
        <div class="panel result-panel">
          <div class="panel-header">
            <ThunderboltOutlined />
            <span>Result</span>
          </div>

          <div class="result-body">
            <!-- Loading -->
            <div v-if="generating" class="result-loading">
              <a-spin size="large" />
              <p>Generating...</p>
            </div>

            <!-- Error -->
            <a-alert
              v-else-if="error"
              type="error"
              :message="error"
              show-icon
              closable
              @close="error = null"
            />

            <!-- Result -->
            <div v-else-if="result" class="result-content">
              <div class="result-meta">
                <a-tag color="green">{{ result.providerName }}</a-tag>
                <a-tag color="blue">{{ result.modelName }}</a-tag>
                <a-tag>{{ formatDuration(result.durationMs) }}</a-tag>
              </div>

              <div v-if="result.data" class="result-media">
                <img
                  v-if="result.mimeType.startsWith('image/')"
                  :src="`data:${result.mimeType};base64,${result.data}`"
                  class="result-image"
                />
                <video
                  v-else-if="result.mimeType.startsWith('video/')"
                  :src="`data:${result.mimeType};base64,${result.data}`"
                  controls
                  class="result-video"
                />
              </div>
            </div>

            <!-- Empty state -->
            <div v-else class="result-empty">
              <ExperimentOutlined :style="{ fontSize: '48px', color: '#CBD5E1' }" />
              <p>Enter a prompt and click Generate to see results here</p>
            </div>
          </div>
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;
}

.header-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: #00D9A515;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #00D9A5;
  font-size: 20px;
  flex-shrink: 0;
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

.panel {
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  border-radius: 14px;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 20px;
  background: #F8FAFC;
  border-bottom: 1px solid #E2E8F0;
  font-size: 14px;
  font-weight: 600;
  color: #0F172A;
}

.gen-form {
  padding: 20px;
}

/* Reference image upload */
.ref-upload {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  border: 2px dashed #E2E8F0;
  border-radius: 10px;
  cursor: pointer;
  color: #94A3B8;
  font-size: 13px;
  transition: border-color 0.2s, background 0.2s;
}

.ref-upload:hover {
  border-color: #00D9A5;
  background: #F0FDF9;
}

.ref-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  background: #F8FAFC;
}

.ref-thumb {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid #E2E8F0;
  flex-shrink: 0;
}

.ref-info {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-width: 0;
}

.ref-name {
  font-size: 13px;
  color: #475569;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.param-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.param-item {
  min-width: 140px;
  flex: 1;
}

.param-label {
  display: block;
  font-size: 12px;
  color: #64748B;
  margin-bottom: 4px;
  font-weight: 500;
}

.result-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.result-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  min-height: 300px;
}

.result-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #64748B;
}

.result-content {
  width: 100%;
}

.result-meta {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.result-media {
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #E2E8F0;
}

.result-image {
  width: 100%;
  max-height: 480px;
  object-fit: contain;
  display: block;
  background: #F8FAFC;
}

.result-video {
  width: 100%;
  max-height: 480px;
  display: block;
  background: #000;
}

.result-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #94A3B8;
  text-align: center;
}

@media (max-width: 768px) {
  .page-title {
    font-size: 24px;
  }
}
</style>
