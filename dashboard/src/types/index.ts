export interface Prompt {
  chatId: number
  templateId: number | null
  templateName: string
  provider: string
  modelName: string
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

export interface ProviderInfo {
  name: string
  models: string[]
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
