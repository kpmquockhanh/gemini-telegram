import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Template, TemplateFormData } from '@/types'
import { listTemplates, createTemplate, updateTemplate, deleteTemplate } from '@/api/client'

export const useTemplatesStore = defineStore('templates', () => {
  const items = ref<Template[]>([])
  const loading = ref(false)

  async function fetchTemplates() {
    loading.value = true
    try {
      const res = await listTemplates()
      items.value = res
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch templates'
      console.error(message, err)
    } finally {
      loading.value = false
    }
  }

  async function addTemplate(data: TemplateFormData) {
    try {
      await createTemplate(data)
      await fetchTemplates()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create template'
      console.error(message, err)
      return false
    }
  }

  async function editTemplate(id: number, data: TemplateFormData) {
    try {
      await updateTemplate(id, data)
      await fetchTemplates()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update template'
      console.error(message, err)
      return false
    }
  }

  async function removeTemplate(id: number) {
    try {
      await deleteTemplate(id)
      await fetchTemplates()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete template'
      console.error(message, err)
      return false
    }
  }

  return {
    items,
    loading,
    fetchTemplates,
    addTemplate,
    editTemplate,
    removeTemplate,
  }
})
