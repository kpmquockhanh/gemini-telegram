<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useTemplatesStore } from '@/stores/templates'
import { message } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import type { Template, TemplateFormData } from '@/types'

const templatesStore = useTemplatesStore()

const editModalVisible = ref(false)
const isEditing = ref(false)
const editId = ref<number | null>(null)
const editForm = ref<TemplateFormData>({
  name: '',
  description: '',
  imagePrompt: '',
  videoPrompt: '',
})

onMounted(() => {
  templatesStore.fetchTemplates()
})

function openCreateModal() {
  isEditing.value = false
  editId.value = null
  editForm.value = {
    name: '',
    description: '',
    imagePrompt: '',
    videoPrompt: '',
  }
  editModalVisible.value = true
}

function openEditModal(template: Template) {
  isEditing.value = true
  editId.value = template.id
  editForm.value = {
    name: template.name,
    description: template.description,
    imagePrompt: template.imagePrompt,
    videoPrompt: template.videoPrompt,
  }
  editModalVisible.value = true
}

async function handleSave() {
  if (!editForm.value.name.trim()) {
    message.error('Name is required')
    return
  }

  let success: boolean
  if (isEditing.value && editId.value !== null) {
    success = await templatesStore.editTemplate(editId.value, editForm.value)
  } else {
    success = await templatesStore.addTemplate(editForm.value)
  }

  if (success) {
    message.success(isEditing.value ? 'Template updated' : 'Template created')
    editModalVisible.value = false
  } else {
    message.error('Failed to save template')
  }
}

async function handleDelete(id: number) {
  const success = await templatesStore.removeTemplate(id)
  if (success) {
    message.success('Template deleted')
  } else {
    message.error('Failed to delete template')
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Templates</h1>
        <p class="page-subtitle">Create and manage reusable prompt templates</p>
      </div>
    </div>

    <div class="toolbar">
      <a-button type="primary" @click="openCreateModal">
        <PlusOutlined />
        Create Template
      </a-button>
      <a-button class="refresh-btn" @click="templatesStore.fetchTemplates()">
        <ReloadOutlined />
        Refresh
      </a-button>
    </div>

    <a-table :data-source="templatesStore.items" :loading="templatesStore.loading" :pagination="false" row-key="id"
      class="data-table">
      <a-table-column title="Name" data-index="name" key="name">
        <template #default="{ text }">
          <span class="template-name">{{ text }}</span>
        </template>
      </a-table-column>
      <a-table-column title="Description" data-index="description" key="description" ellipsis>
        <template #default="{ text }">
          <span v-if="text" class="desc-text">{{ text }}</span>
          <span v-else class="empty-text">—</span>
        </template>
      </a-table-column>
      <a-table-column title="Image Prompt" data-index="imagePrompt" key="imagePrompt" ellipsis>
        <template #default="{ text }">
          <span v-if="text" class="prompt-text">{{ text }}</span>
          <span v-else class="empty-text">—</span>
        </template>
      </a-table-column>
      <a-table-column title="Video Prompt" data-index="videoPrompt" key="videoPrompt" ellipsis>
        <template #default="{ text }">
          <span v-if="text" class="prompt-text">{{ text }}</span>
          <span v-else class="empty-text">—</span>
        </template>
      </a-table-column>
      <a-table-column title="Actions" key="actions" width="120" align="center">
        <template #default="{ record }">
          <div class="action-btns">
            <a-button type="text" class="action-btn edit" size="small" @click="openEditModal(record)">
              <EditOutlined />
            </a-button>
            <a-popconfirm title="Delete this template?" @confirm="handleDelete(record.id)">
              <a-button type="text" class="action-btn delete" size="small">
                <DeleteOutlined />
              </a-button>
            </a-popconfirm>
          </div>
        </template>
      </a-table-column>
    </a-table>

    <!-- Edit/Create Modal -->
    <a-modal v-model:open="editModalVisible" :title="isEditing ? 'Edit Template' : 'Create Template'" ok-text="Save"
      cancel-text="Cancel" class="custom-modal" @ok="handleSave">
      <a-form layout="vertical">
        <a-form-item label="Name" required>
          <a-input v-model:value="editForm.name" placeholder="Template name" />
        </a-form-item>

        <a-form-item label="Description">
          <a-input v-model:value="editForm.description" placeholder="Brief description" />
        </a-form-item>

        <a-form-item label="Image Prompt">
          <a-textarea v-model:value="editForm.imagePrompt" :rows="4" :maxlength="2000" show-count
            placeholder="Default image prompt..." />
        </a-form-item>

        <a-form-item label="Video Prompt">
          <a-textarea v-model:value="editForm.videoPrompt" :rows="4" :maxlength="2000" show-count
            placeholder="Default video prompt..." />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 28px;
  font-weight: 700;
  color: #0F172A;
  margin: 0;
  letter-spacing: -0.03em;
}

.page-subtitle {
  color: #475569;
  margin: 6px 0 0;
  font-size: 15px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
}

.refresh-btn {
  color: #475569;
}

.data-table {
  border-radius: 14px;
  overflow: hidden;
}

.template-name {
  font-weight: 600;
  color: #0F172A;
}

.desc-text {
  color: #475569;
}

.prompt-text {
  color: #0F172A;
}

.empty-text {
  color: #94A3B8;
  font-style: italic;
}

.action-btns {
  display: flex;
  gap: 4px;
  justify-content: center;
}

.action-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.15s;
}

.action-btn.edit {
  color: #00D9A5;
}

.action-btn.edit:hover {
  background: rgba(0, 217, 165, 0.1);
}

.action-btn.delete {
  color: #EF4444;
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.1);
}

/* Responsive */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
