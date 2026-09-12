import { createRouter, createWebHistory } from 'vue-router'

import { useStore } from '@/store'
import { PermissionType } from '@/contracts/users/request.schema'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/pages/landing/LandingPage.vue'),
      meta: {
        layout_0: 'FooterBlockLayout',
      },
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
    {
      path: '/basket',
      meta: {
        layer_0: 'BaseFooterLayout',
        layer_1: 'HasAccessTokenLayout',
      },
      children: [
        {
          path: '/basket/:id',
          name: 'basket',
          component: () => import('@/pages/basket/BasketPage.vue'),
        },
      ],
    },
    {
      path: '/auth',
      meta: {
        layout_0: 'BaseFooterLayout',
      },
      children: [
        {
          path: '/auth/',
          name: 'auth',
          component: () => import('@/pages/auth/AuthPage.vue'),
        },
        {
          path: '/auth/login',
          name: 'login',
          component: () => import('@/pages/auth/LoginPage.vue'),
        },
        {
          path: '/auth/register',
          name: 'register',
          component: () => import('@/pages/auth/RegisterPage.vue'),
        },
      ],
    },
    {
      path: '/panel',
      meta: {
        layout_0: 'BaseFooterLayout',
        // layout_1: 'HasAccessTokenLayout',
      },
      children: [
        {
          path: '/panel/admin',
          // meta: {
          //   permissions_list: [PermissionType.Admin],
          // },
          children: [
            {
              path: '/panel/admin/',
              name: 'admin',
              component: () => import('@/pages/panel/admin/AdminPage.vue'),
            },
            {
              path: '/panel/admin/users',
              children: [
                {
                  path: '/panel/admin/users/',
                  name: 'admin-users-list',
                  component: () =>
                    import('@/pages/panel/admin/users/UsersListPage.vue'),
                },
                {
                  path: '/panel/admin/users/:id',
                  name: 'admin-user-info',
                  component: () =>
                    import('@/pages/panel/admin/users/UserInfoPage.vue'),
                },
              ],
            },
            {
              path: '/panel/admin/categories',
              name: 'admin-categories',
              component: () =>
                import('@/pages/panel/admin/categories/CategoriesPage.vue'),
            },
            {
              path: '/panel/admin/products',
              children: [
                {
                  path: '/panel/admin/products/',
                  name: 'admin-products-list',
                  component: () =>
                    import('@/pages/panel/admin/products/ProductsListPage.vue'),
                },
                {
                  path: '/panel/admin/products/create',
                  name: 'admin-product-create',
                  component: () =>
                    import('@/pages/panel/admin/products/CreateProductPage.vue'),
                },
                {
                  path: '/panel/admin/products/upload/:productID',
                  name: 'admin-product-upload-img',
                  component: () =>
                    import('@/pages/panel/admin/products/UploadImagePage.vue'),
                },
                {
                  path: '/panel/admin/products/edit',
                  name: 'admin-product-edit',
                  component: () =>
                    import('@/pages/panel/admin/products/EditPage.vue'),
                },
              ],
            },
          ],
        },
      ],
    },
  ],
})

router.onError((error) => {
  console.log(error)
})

router.beforeEach((to) => {
  const store = useStore()
  const permissions_list = to.meta.permissions_list
  if (permissions_list === undefined) {
    return true
  }
  if (!Array.isArray(permissions_list)) {
    console.error('permissions_list must be array of number')
    return { path: '/', replace: true }
  }
  if (!permissions_list.includes(store.getClaims.permission_type)) {
    return { path: '/', replace: true }
  }
  return true
})

export default router
