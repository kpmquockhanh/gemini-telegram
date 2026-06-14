import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import PromptsView from '@/views/PromptsView.vue'
import TemplatesView from '@/views/TemplatesView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: DashboardView,
    },
    {
      path: '/prompts',
      name: 'prompts',
      component: PromptsView,
    },
    {
      path: '/templates',
      name: 'templates',
      component: TemplatesView,
    },
  ],
})

export default router
