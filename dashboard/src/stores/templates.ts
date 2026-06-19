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
    const start = performance.now()
    loading.value = true
    try {
      const res = await listTemplates()
      items.value = res
      logs.addGeneration({
        operation: 'listTemplates',
        status: 'success',
        params: {},
        result: { count: res.length },
        durationMs: Math.round(performance.now() - start),
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch templates'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'listTemplates' } })
      logs.addGeneration({
        operation: 'listTemplates',
        status: 'error',
        params: {},
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
    } finally {
      loading.value = false
    }
  }

  async function addTemplate(data: TemplateFormData) {
    const logs = useLogsStore()
    const start = performance.now()
    try {
      const res = await createTemplate(data)
      await fetchTemplates()
      logs.addGeneration({
        operation: 'createTemplate',
        status: 'success',
        params: data as unknown as Record<string, unknown>,
        result: res,
        durationMs: Math.round(performance.now() - start),
      })
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create template'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'createTemplate', ...data } })
      logs.addGeneration({
        operation: 'createTemplate',
        status: 'error',
        params: data as unknown as Record<string, unknown>,
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
      return false
    }
  }

  async function editTemplate(id: number, data: TemplateFormData) {
    const logs = useLogsStore()
    const start = performance.now()
    try {
      const res = await updateTemplate(id, data)
      await fetchTemplates()
      logs.addGeneration({
        operation: 'updateTemplate',
        status: 'success',
        params: { id, ...data } as unknown as Record<string, unknown>,
        result: res,
        durationMs: Math.round(performance.now() - start),
      })
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update template'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'updateTemplate', id, ...data } })
      logs.addGeneration({
        operation: 'updateTemplate',
        status: 'error',
        params: { id, ...data } as unknown as Record<string, unknown>,
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
      return false
    }
  }

  async function removeTemplate(id: number) {
    const logs = useLogsStore()
    const start = performance.now()
    try {
      const res = await deleteTemplate(id)
      await fetchTemplates()
      logs.addGeneration({
        operation: 'deleteTemplate',
        status: 'success',
        params: { id },
        result: res,
        durationMs: Math.round(performance.now() - start),
      })
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete template'
      console.error(message, err)
      logs.addTraceLog({ level: 'error', message, context: { action: 'deleteTemplate', id } })
      logs.addGeneration({
        operation: 'deleteTemplate',
        status: 'error',
        params: { id },
        result: { error: message },
        durationMs: Math.round(performance.now() - start),
      })
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
