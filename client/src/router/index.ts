import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/pages/landing/LandingPage.vue'),
    },
    {
      path: '/products',
      children: [
        {
          path: '/products/',
          name: 'products-list',
          component: () => import('@/pages/products/ProductsListPage.vue'),
        },
        {
          path: '/products/:id',
          name: 'product',
          component: () => import('@/pages/products/ProductPage.vue'),
        },
      ],
    },
  ],
})

export default router
