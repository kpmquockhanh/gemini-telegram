export interface Prompt {
  chatId: number
  templateId: number | null
  templateName: string
  provider: string
  modelName: string
  params: string
  updatedAt: string
}

export interface PromptListResponse {
  items: Prompt[]
  total: number
  page: number
  limit: number
}

export interface PromptFormData {
  templateId: number | null
  provider: string
  modelName: string
  params: Record<string, unknown>
}

export interface Template {
  id: number
  name: string
  description: string
  imagePrompt: string
  videoPrompt: string
  createdAt: string
}

export interface TemplateFormData {
  name: string
  description: string
  imagePrompt: string
  videoPrompt: string
}

export interface ModelParamOption {
  label: string
  value: string
}

export interface ModelParamDef {
  name: string
  label: string
  type: 'slider' | 'number' | 'select'
  default: unknown
  min?: number
  max?: number
  step?: number
  options?: ModelParamOption[]
}

export interface ModelInfo {
  name: string
  params: ModelParamDef[]
}

export interface ProviderInfo {
  name: string
  models: ModelInfo[]
}

export interface ProvidersResponse {
  providers: Record<string, ProviderInfo>
  default: string
}

export interface Stats {
  totalChats: number
  templateCount: number
  totalPrompts: number
}

export interface GenerationHistory {
  id: number
  chatId: number
  jobType: string
  prompt: string
  provider: string
  modelName: string
  status: 'success' | 'error'
  params: string
  result: string
  durationMs: number
  createdAt: string
}

export interface GenerationHistoryListResponse {
  items: GenerationHistory[]
  total: number
  page: number
  limit: number
}

export interface GenerateRequest {
  type: 'image' | 'video'
  prompt: string
  provider: string
  modelName: string
  params?: Record<string, unknown>
  referenceImage?: string
  referenceMimeType?: string
}

export interface GenerateResponse {
  data: string
  mimeType: string
  providerName: string
  modelName: string
  durationMs: number
  error?: string
}
