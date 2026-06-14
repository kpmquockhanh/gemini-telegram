export interface Prompt {
  chatId: number
  imagePrompt: string
  videoPrompt: string
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
  imagePrompt: string
  videoPrompt: string
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
  imageCount: number
  videoCount: number
  totalPrompts: number
}
