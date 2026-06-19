import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Template, TemplateFormData } from '@/types'
import { listTemplates, createTemplate, updateTemplate, deleteTemplate } from '@/api/client'
import { useLogsStore } from '@/stores/logs'

export const useTemplatesStore = defineStore('templates', () => {
  const items = ref<Template[]>([])
  const loading = ref(false)

  async function fetchTemplates() {
    const logs = useLogsStore()
    loading.value = true
    try {
      const res = await listTemplates()
      items.value = res
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch templates'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'listTemplates' } })
    } finally {
      loading.value = false
    }
  }

  async function addTemplate(data: TemplateFormData) {
    const logs = useLogsStore()
    try {
      await createTemplate(data)
      await fetchTemplates()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create template'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'createTemplate', ...data } })
      return false
    }
  }

  async function editTemplate(id: number, data: TemplateFormData) {
    const logs = useLogsStore()
    try {
      await updateTemplate(id, data)
      await fetchTemplates()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update template'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'updateTemplate', id, ...data } })
      return false
    }
  }

  async function removeTemplate(id: number) {
    const logs = useLogsStore()
    try {
      await deleteTemplate(id)
      await fetchTemplates()
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete template'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'deleteTemplate', id } })
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
