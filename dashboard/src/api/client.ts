import axios from 'axios'
import type { Prompt, PromptListResponse, PromptFormData, Template, TemplateFormData, Stats, ProvidersResponse, GenerationHistoryListResponse, GenerateRequest, GenerateResponse } from '@/types'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Prompts
export const listPrompts = (page = 1, limit = 20, search = ''): Promise<PromptListResponse> =>
  client.get('/prompts', { params: { page, limit, search } }).then((r) => r.data)

export const getPrompt = (chatId: number): Promise<Prompt> =>
  client.get(`/prompts/${chatId}`).then((r) => r.data)

export const updatePrompt = (chatId: number, data: PromptFormData): Promise<Prompt> =>
  client.put(`/prompts/${chatId}`, data).then((r) => r.data)

export const deletePrompt = (chatId: number): Promise<{ message: string }> =>
  client.delete(`/prompts/${chatId}`).then((r) => r.data)

// Stats
export const getStats = (): Promise<Stats> =>
  client.get('/stats').then((r) => r.data)

// Providers
export const getProviders = (): Promise<ProvidersResponse> =>
  client.get('/providers').then((r) => r.data)

// Templates
export const listTemplates = (): Promise<Template[]> =>
  client.get('/templates').then((r) => r.data)

export const createTemplate = (data: TemplateFormData): Promise<{ id: number }> =>
  client.post('/templates', data).then((r) => r.data)

export const updateTemplate = (id: number, data: TemplateFormData): Promise<{ message: string }> =>
  client.put(`/templates/${id}`, data).then((r) => r.data)

export const deleteTemplate = (id: number): Promise<{ message: string }> =>
  client.delete(`/templates/${id}`).then((r) => r.data)

// Generation History
export const listGenerationHistory = (page = 1, limit = 20, status = 'all', search = ''): Promise<GenerationHistoryListResponse> =>
  client.get('/generations', { params: { page, limit, status, search } }).then((r) => r.data)

export const clearGenerationHistory = (): Promise<{ message: string }> =>
  client.delete('/generations').then((r) => r.data)

// Generate
export const generate = (data: GenerateRequest): Promise<GenerateResponse> =>
  client.post('/generate', data).then((r) => r.data)
