import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Prompt, PromptListResponse, PromptFormData } from '@/types'
import { listPrompts, updatePrompt, deletePrompt } from '@/api/client'

export const usePromptsStore = defineStore('prompts', () => {
  const items = ref<Prompt[]>([])
  const total = ref(0)
  const loading = ref(false)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const searchQuery = ref('')

  async function fetchPrompts(page = currentPage.value, limit = pageSize.value, search = searchQuery.value) {
    loading.value = true
    try {
      const res: PromptListResponse = await listPrompts(page, limit, search)
      items.value = res.items
      total.value = res.total
      currentPage.value = page
      pageSize.value = limit
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch prompts'
      console.error(message, err)
    } finally {
      loading.value = false
    }
  }

  async function savePrompt(chatId: number, data: PromptFormData) {
    try {
      await updatePrompt(chatId, data)
      await fetchPrompts()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to save prompt'
      console.error(message, err)
      return false
    }
  }

  async function removePrompt(chatId: number) {
    try {
      await deletePrompt(chatId)
      await fetchPrompts()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete prompt'
      console.error(message, err)
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
