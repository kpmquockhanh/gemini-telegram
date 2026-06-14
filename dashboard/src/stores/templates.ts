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
      items.value = await listTemplates()
    } catch (err) {
      console.error('Failed to fetch templates', err)
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
      console.error('Failed to create template', err)
      return false
    }
  }

  async function editTemplate(id: number, data: TemplateFormData) {
    try {
      await updateTemplate(id, data)
      await fetchTemplates()
      return true
    } catch (err) {
      console.error('Failed to update template', err)
      return false
    }
  }

  async function removeTemplate(id: number) {
    try {
      await deleteTemplate(id)
      await fetchTemplates()
      return true
    } catch (err) {
      console.error('Failed to delete template', err)
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
