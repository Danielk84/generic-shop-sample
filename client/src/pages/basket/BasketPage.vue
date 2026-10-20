<script setup lang="ts">
import { computed, defineAsyncComponent, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useNotificationStore } from '@/store/notification'
import type {
  OwnedOrderItemResponse,
  OrderResponse,
} from '@/contracts/orders/response.interface'

const OrderItems = defineAsyncComponent(
  () => import('@/pages/basket/OrderItems.vue'),
)
const CardTotal = defineAsyncComponent(
  () => import('@/pages/basket/CartTotal.vue'),
)
const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const route = useRoute()
const router = useRouter()
const store = useStore()
const notification = useNotificationStore()

const orderId = computed(() => String(route.params.id ?? ''))
const page = ref<number>(1)
const maxPage = ref<number>(1)

const {
  data: order,
  isPending: isOrderPending,
  error: orderError,
  refetch: refetchOrder,
} = useQuery({
  queryKey: computed(() => ['basket-order', orderId.value]),
  queryFn: async () =>
    api.get<OrderResponse>(`orders/${orderId.value}`, {
      headers: { Authorization: store.getAccessToken },
    }),
  select: (res) => res.data,
  enabled: computed(() => orderId.value !== '' && orderId.value !== '0'),
})

watch(orderError, (err) => {
  if (err !== null) {
    errorStatusHandler(err, router, {
      notFound() {
        // empty block
        return
      },
    })
  }
})

// Local reactive state for basket items
const basketItems = ref<
  Array<{
    product_id: string
    img: string | undefined
    alt: string
    name: string
    price: number
    count: number
  }>
>([])

const {
  data,
  error,
  isPending: isItemsPending,
  refetch,
} = useQuery({
  queryKey: computed(() => ['basket-items', orderId.value, page.value]),
  queryFn: async () =>
    api.get<OwnedOrderItemResponse[]>(
      `orders/items/customer/${orderId.value}`,
      {
        params: { page: page.value },
        headers: { Authorization: store.getAccessToken },
      },
    ),
  select: (res) => {
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
    return res.data
  },
  enabled: computed(() => orderId.value !== '' && orderId.value !== '0'),
})

watch(
  data,
  (newData) => {
    if (newData) {
      basketItems.value = newData.map((item) => ({
        product_id: item.product_id,
        img: undefined,
        alt: item.name,
        name: item.name,
        price: item.price,
        count: item.items_total || 1,
      }))
    }
  },
  { immediate: true },
)

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          // empty block
          return
        },
      })
    }
  },
  { immediate: true },
)

const items = computed(() =>
  basketItems.value.map((item) => ({
    ...item,
    total: item.price * item.count,
  })),
)

// Totals are read from the order so paginating the item list never truncates
// the item count or the total bill.
const totalItems = computed(() => order.value?.items_total ?? 0)
const itemsTotal = computed(() => order.value?.total_bill ?? 0)

const isLoading = computed(() => isOrderPending.value || isItemsPending.value)

watch(
  isLoading,
  (value) => {
    store.setLoadingStatus(value)
  },
  { immediate: true },
)

onUnmounted(() => store.setLoadingStatus(false))

const removeMutation = useMutation({
  mutationKey: ['basket-remove'],
  mutationFn: async (input: { order_id: string; product_id: string }) => {
    return api.delete('orders/items/', {
      data: input,
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: async () => {
    notification.success('Item removed.')
    await Promise.all([refetch(), refetchOrder()])
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
})

const changeCountMutation = useMutation({
  mutationKey: ['basket-change-count'],
  mutationFn: async (input: {
    order_id: string
    product_id: string
    total: number
  }) => {
    await api.put(
      `orders/items/set-items-total/${input.total}`,
      { order_id: input.order_id, product_id: input.product_id },
      { headers: { Authorization: store.getAccessToken } },
    )
  },
  onSuccess: async () => {
    notification.success('Count updated.')
    await Promise.all([refetch(), refetchOrder()])
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
})

const processOrderMutation = useMutation({
  mutationKey: ['basket-process-order'],
  mutationFn: async () => {
    return api.put(
      `orders/verify/${orderId.value}`,
      { accepted: true },
      { headers: { Authorization: store.getAccessToken } },
    )
  },
  onSuccess: async () => {
    notification.success('Order verified. Continue to shipping info.')
    router.push({ name: 'order-detail', params: { id: orderId.value } })
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
})

function handleChangeCount(input: { productId: string; count: number }) {
  const item = basketItems.value.find((i) => i.product_id === input.productId)
  if (item) {
    item.count = input.count
  }
  changeCountMutation.mutate({
    order_id: orderId.value,
    product_id: input.productId,
    total: input.count,
  })
}

function handleRemove(productId: string) {
  basketItems.value = basketItems.value.filter(
    (i) => i.product_id !== productId,
  )
  removeMutation.mutate({
    order_id: orderId.value,
    product_id: productId,
  })
}

function handleProcessOrder() {
  if (orderId.value === '' || orderId.value === '0') {
    return
  }
  if (totalItems.value === 0) {
    notification.warning('Your basket is empty.')
    return
  }
  processOrderMutation.mutate()
}
</script>

<template>
  <div class="basket-page">
    <OrderItems
      :orders="items"
      @remove="handleRemove"
      @change-count="handleChangeCount"
    />
    <CardTotal
      :data="{
        totalItem: totalItems,
        total: itemsTotal,
      }"
      :disabled="totalItems === 0"
      :processing="processOrderMutation.isPending.value"
      @process="handleProcessOrder"
    />
    <ListPagination
      :last="maxPage"
      page-name="basket"
      @change-page="
        (newValue: number) => {
          page = newValue
        }
      "
    />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.basket-page {
  @apply w-full min-h-screen p-4 md:p-10
    flex flex-col items-center gap-6 md:gap-10;
}
</style>
