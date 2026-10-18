<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler, formatDate } from '@/utils/helper'
import type { OrderSummaryResponse } from '@/contracts/orders/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const route = useRoute()
const router = useRouter()
const store = useStore()

const page = ref<number>(1)
const maxPage = ref<number>(1)

const { data, error } = useQuery({
  queryKey: computed(() => ['orders-customer', page.value]),
  queryFn: async () =>
    api.get<OrderSummaryResponse[]>('orders/customer', {
      params: { page: page.value },
      headers: { Authorization: store.getAccessToken },
    }),
  select: (res) => {
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
    return res.data
  },
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          const p = Number(route.query?.page)
          if (p > 1) {
            router.back()
          }
        },
      })
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="orders-list-page">
    <h1 class="page-title">My Orders</h1>
    <div v-if="data === undefined || data.length === 0" class="empty">
      <p>You have no orders yet.</p>
    </div>
    <div v-else class="orders">
      <RouterLink
        v-for="item in data"
        :key="item.id"
        class="order-item"
        :to="{ name: 'order-detail', params: { id: item.id } }"
      >
        <div class="row">
          <span class="label">Started</span>
          <span>{{ formatDate(item.started_at) }}</span>
        </div>
        <div class="badges">
          <span class="badge" :class="item.is_paid ? 'ok' : 'bad'">
            {{ item.is_paid ? 'Paid' : 'Unpaid' }}
          </span>
          <span class="badge" :class="item.is_delivered ? 'ok' : 'bad'">
            {{ item.is_delivered ? 'Delivered' : 'Pending' }}
          </span>
        </div>
      </RouterLink>
    </div>
    <ListPagination
      :last="maxPage"
      page-name="orders-list"
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

.orders-list-page {
  @apply w-full min-h-screen p-4 md:p-10
    flex flex-col items-center gap-6 md:gap-8;
}

.page-title {
  @apply text-2xl md:text-3xl font-bold;
}

.empty {
  @apply text-xl md:text-2xl font-bold p-6 md:p-10;
}

.orders {
  @apply w-full max-w-200 flex flex-col gap-4;
}

.order-item {
  @apply w-full p-4 md:p-5 rounded-2xl
    border border-(--c-v-5)
    flex flex-row justify-between items-center gap-4
    sm:flex-row sm:items-center
    hover:shadow-[0_0_15px_4px_var(--c-v-9)]
    transition duration-150;
}

.row {
  @apply flex flex-col gap-1;
}

.label {
  @apply text-sm text-(--c-v-4);
}

.badges {
  @apply flex flex-row gap-2;
}

.badge {
  @apply px-3 py-1 rounded-xl text-sm font-bold;
}

.badge.ok {
  @apply bg-(--c-v-14) text-(--c-v-1);
}

.badge.bad {
  @apply bg-(--c-v-11) text-(--c-v-1);
}

@media (max-width: 640px) {
  .order-item {
    @apply flex-col items-start;
  }

  .badges {
    @apply w-full justify-between;
  }
}
</style>
