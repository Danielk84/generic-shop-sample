<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { DailySalesResponse } from '@/contracts/orders/response.interface'

const OrderListView = defineAsyncComponent(
  () => import('@/components/common/orders/OrderListView.vue'),
)

const router = useRouter()
const store = useStore()

const orderId = ref<string>('')
const isSearchTriggerd = ref<boolean>(false)

const dailySales = ref<DailySalesResponse[]>([])
const dailySalesPage = ref<number>(1)

const dailySalesQuery = useQuery({
  queryKey: computed(() => ['admin-daily-sales', dailySalesPage.value]),
  queryFn: async () =>
    api.get<DailySalesResponse[]>('orders/items/daily-sales', {
      params: { page: dailySalesPage.value },
      headers: { Authorization: store.getAccessToken },
    }),
  select: (res) => res.data,
})

watch(dailySalesQuery.data, (value) => {
  if (value !== undefined) {
    dailySales.value = value
  }
})

watch(dailySalesQuery.error, (error) => {
  if (error !== undefined) {
    errorStatusHandler(error, router, {
      notFound() {
        // empty block
        return
      },
    })
  }
})

function findItems() {
  orderId.value = orderId.value.trim()
  isSearchTriggerd.value = true
}
</script>

<template>
  <div class="admin-orders-page">
    <h1>Order Administration</h1>

    <section class="card">
      <h2>Daily sales</h2>
      <div v-if="dailySales.length === 0" class="empty">
        <p>No sales data.</p>
      </div>
      <div v-else class="w-full c-flex-all-center gap-2 flex-col">
        <div v-for="sale in dailySales" :key="sale.date" class="row">
          <span>{{ sale.date }}</span>
          <strong>{{ sale.count }}</strong>
        </div>
      </div>
    </section>

    <section class="card">
      <h2>Order items</h2>
      <form class="finder" @submit.prevent="findItems">
        <label for="order-id">Order ID</label>
        <div class="search">
          <input
            class="c-clean-input"
            v-model="orderId"
            type="text"
            placeholder="Enter order UUID"
          />
        </div>
        <button type="submit">Find items</button>
      </form>
      <section v-if="isSearchTriggerd" class="w-full">
        <OrderListView :orderID="orderId" target="admin" />
      </section>
    </section>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.admin-orders-page {
  @apply w-full min-h-screen p-4 md:p-10 flex flex-col items-center gap-6 md:gap-8;
}

.admin-orders-page > h1 {
  @apply text-2xl md:text-3xl font-bold;
}

.card {
  @apply w-full max-w-300 p-4 md:p-6 rounded-2xl border-2 border-(--c-v-15)
    flex flex-col gap-4;
}

.card h2 {
  @apply text-xl md:text-2xl font-bold;
}

.card .empty {
  @apply w-full text-lg text-center;
}

.card .row {
  @apply p-4 w-full flex items-center justify-between
    bg-(--c-v-5) dark:bg-(--c-v-3) rounded-2xl;
}

.card .finder {
  @apply flex items-center w-full flex-wrap gap-4;
}

.card .search {
  @apply w-full max-w-120 min-w-60
    border-3 rounded-2xl border-(--c-v-15)
    p-4 text-xl;
}

.finder button {
  @apply cursor-pointer bg-(--c-v-15)
    rounded-2xl text-(--c-v-7) p-4 font-bold;
}
</style>
