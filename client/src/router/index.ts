import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/components/pages/HomePage.vue'),
    },
    {
      path: "/auth",
      name: "auth",
      component: () => import('@/components/pages/AuthPage.vue'),
    },
  ],
})

export default router
