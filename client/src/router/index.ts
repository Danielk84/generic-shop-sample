import { createRouter, createWebHistory } from 'vue-router'

import { useStore } from '@/store'
import { PermissionType } from '@/contracts/users/request.schema'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior() {
    return {
      top: 0,
      left: 0,
    }
  },
  routes: [
    {
      path: '/',
      meta: {
        layout_0: 'FooterBlockLayout',
      },
      children: [
        {
          path: '/',
          name: 'home',
          component: () => import('@/pages/landing/LandingPage.vue'),
        },
        {
          path: '/categories',
          name: 'categories',
          component: () => import('@/pages/products/CategoriesPage.vue'),
        },
        {
          path: '/400',
          name: 'error-bad-request',
          component: () => import('@/pages/error/BadRequestPage.vue'),
        },
        {
          path: '/401',
          name: 'error-unauthorized',
          component: () => import('@/pages/error/UnauthorizedPage.vue'),
        },
        {
          path: '/403',
          name: 'error-forbidden',
          component: () => import('@/pages/error/ForbiddenPage.vue'),
        },
        {
          path: '/404',
          name: 'error-not-found',
          component: () => import('@/pages/error/NotFoundPage.vue'),
        },
        {
          path: '/406',
          name: 'error-not-acceptable',
          component: () => import('@/pages/error/NotAcceptablePage.vue'),
        },
        {
          path: '/413',
          name: 'error-content-too-large',
          component: () => import('@/pages/error/ContentTooLarge.vue'),
        },
        {
          path: '/422',
          name: 'error-unprocessable-content',
          component: () => import('@/pages/error/UnprocessableContentPage.vue'),
        },
        {
          path: '/429',
          name: 'error-too-many-requests',
          component: () => import('@/pages/error/TooManyRequestsPage.vue'),
        },
        {
          path: '/500',
          name: 'error-internal-server-error',
          component: () => import('@/pages/error/InternalServerErrorPage.vue'),
        },
      ],
    },
    {
      path: '/products',
      meta: {
        layout_0: 'FooterBlockLayout',
      },
      children: [
        {
          path: '/products/',
          name: 'products-list',
          component: () => import('@/pages/products/ProductsListPage.vue'),
        },
        {
          path: '/products/popular',
          name: 'products-popular',
          component: () => import('@/pages/products/ProductsListPage.vue'),
        },
        {
          path: '/products/:productID',
          name: 'product',
          component: () =>
            import('@/components/common/products/ProductView.vue'),
          props: {
            queryKey: 'public-product',
            queryPath: '/products',
          },
        },
      ],
    },
    {
      path: '/basket',
      meta: {
        layout_0: 'BaseFooterLayout',
        layout_1: 'HasAccessTokenLayout',
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
      path: '/orders',
      meta: {
        layout_0: 'BaseFooterLayout',
        layout_1: 'HasAccessTokenLayout',
      },
      children: [
        {
          path: '/orders/',
          name: 'orders-list',
          component: () => import('@/pages/panel/orders/OrdersListPage.vue'),
        },
        {
          path: '/orders/:id',
          name: 'order-detail',
          component: () => import('@/pages/panel/orders/OrderPage.vue'),
        },
        {
          path: '/orders/verify-payment',
          name: 'user-verify-payment',
          component: () => import('@/pages/panel/orders/VerifyPaymentPage.vue'),
        },
      ],
    },
    {
      path: '/account',
      meta: {
        layout_0: 'BaseFooterLayout',
        layout_1: 'HasAccessTokenLayout',
      },
      children: [
        {
          path: '/account/',
          name: 'account',
          component: () => import('@/pages/account/AccountPage.vue'),
        },
      ],
    },
    {
      path: '/panel',
      meta: {
        layout_0: 'BaseFooterLayout',
        layout_1: 'HasAccessTokenLayout',
      },
      children: [
        {
          path: '/panel/vendors',
          name: 'vendor-orders',
          meta: {
            permissions_list: [PermissionType.Admin, PermissionType.Vendor],
          },
          component: () => import('@/pages/panel/vendors/VendorOrdersPage.vue'),
        },
        {
          path: '/panel/issues',
          name: 'issues-list',
          meta: {
            permissions_list: [
              PermissionType.Customer,
              PermissionType.Vendor,
              PermissionType.Admin,
            ],
          },
          component: () => import('@/pages/panel/issues/IssuesListPage.vue'),
        },
        {
          path: '/panel/admin',
          meta: {
            permissions_list: [PermissionType.Admin],
          },
          children: [
            {
              path: '/panel/admin/',
              name: 'admin',
              component: () => import('@/pages/panel/admin/AdminPage.vue'),
            },
            {
              path: '/panel/admin/orders',
              name: 'admin-orders',
              component: () =>
                import('@/pages/panel/admin/orders/AdminOrdersPage.vue'),
            },
            {
              path: '/panel/admin/issues',
              name: 'admin-issues',
              component: () =>
                import('@/pages/panel/admin/AdminIssuesListPage.vue'),
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
                  path: '/panel/admin/products/edit/:productID',
                  name: 'admin-product-edit',
                  component: () =>
                    import('@/pages/panel/admin/products/EditPage.vue'),
                },
                {
                  path: '/panel/admin/products/overview/:productID',
                  name: 'admin-product-overview',
                  component: () =>
                    import('@/components/common/products/ProductView.vue'),
                  props: {
                    queryKey: 'admin-overview-product',
                    queryPath: 'products/overview',
                  },
                },
              ],
            },
            {
              path: '/panel/admin/comments',
              children: [
                {
                  path: '/panel/admin/comments/',
                  name: 'admin-comments-list',
                  component: () =>
                    import('@/pages/panel/admin/comments/CommentsList.vue'),
                },
              ],
            },
            {
              path: '/panel/admin/shops',
              children: [
                {
                  path: '/panel/admin/shops',
                  name: 'admin-shops-list',
                  component: () =>
                    import('@/pages/panel/admin/shops/ShopsListPage.vue'),
                },
                {
                  path: '/panel/admin/shops/:id',
                  name: 'admin-shops-detail',
                  component: () =>
                    import('@/pages/panel/admin/shops/ShopDetailPage.vue'),
                },
              ],
            },
          ],
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: {
        name: 'error-not-found',
      },
    },
  ],
})

router.onError((error) => {
  console.error(error)
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
