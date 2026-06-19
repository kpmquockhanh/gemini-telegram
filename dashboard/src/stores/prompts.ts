import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Prompt, PromptListResponse, PromptFormData } from '@/types'
import { listPrompts, updatePrompt, deletePrompt } from '@/api/client'
import { useLogsStore } from '@/stores/logs'

export const usePromptsStore = defineStore('prompts', () => {
  const items = ref<Prompt[]>([])
  const total = ref(0)
  const loading = ref(false)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const searchQuery = ref('')

  async function fetchPrompts(page = currentPage.value, limit = pageSize.value, search = searchQuery.value) {
    const logs = useLogsStore()
    const start = performance.now()
    loading.value = true
    try {
      const res: PromptListResponse = await listPrompts(page, limit, search)
      items.value = res.items
      total.value = res.total
      currentPage.value = page
      pageSize.value = limit
      logs.addGeneration({
        operation: 'listPrompts',
        status: 'success',
        params: { page, limit, search },
        result: { total: res.total, count: res.items.length },
        durationMs: Math.round(performance.now() - start),
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch prompts'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'listPrompts', page, limit, search } })
      logs.addGeneration({
        operation: 'listPrompts',
        status: 'error',
        params: { page, limit, search },
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
    } finally {
      loading.value = false
    }
  }

  async function savePrompt(chatId: number, data: PromptFormData) {
    const logs = useLogsStore()
    const start = performance.now()
    try {
      const res = await updatePrompt(chatId, data)
      await fetchPrompts()
      logs.addGeneration({
        operation: 'updatePrompt',
        status: 'success',
        params: { chatId, ...data },
        result: res,
        durationMs: Math.round(performance.now() - start),
      })
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to save prompt'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'updatePrompt', chatId, ...data } })
      logs.addGeneration({
        operation: 'updatePrompt',
        status: 'error',
        params: { chatId, ...data },
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
      return false
    }
  }

  async function removePrompt(chatId: number) {
    const logs = useLogsStore()
    const start = performance.now()
    try {
      const res = await deletePrompt(chatId)
      await fetchPrompts()
      logs.addGeneration({
        operation: 'deletePrompt',
        status: 'success',
        params: { chatId },
        result: res,
        durationMs: Math.round(performance.now() - start),
      })
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete prompt'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'deletePrompt', chatId } })
      logs.addGeneration({
        operation: 'deletePrompt',
        status: 'error',
        params: { chatId },
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
      return false
    }
  }

  return {
    items,
    total,
    loading,
    currentPage,
    pageSize,
    searchQuery,
    fetchPrompts,
    savePrompt,
    removePrompt,
  }
})
