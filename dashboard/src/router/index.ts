import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import PromptsView from '@/views/PromptsView.vue'
import TemplatesView from '@/views/TemplatesView.vue'
import LogsView from '@/views/LogsView.vue'
import GenerationsView from '@/views/GenerationsView.vue'

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
    {
      path: '/logs',
      name: 'logs',
      component: LogsView,
    },
    {
      path: '/generations',
      name: 'generations',
      component: GenerationsView,
    },
  ],
})

export default router
