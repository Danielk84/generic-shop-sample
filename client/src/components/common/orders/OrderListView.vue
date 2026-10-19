<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { OwnedOrderItemResponse } from '@/contracts/orders/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const store = useStore()
const router = useRouter()

const page = ref<number>(1)
const maxPage = ref<number>(1)

const props = defineProps<{
  orderID: string
  target: 'admin' | 'customer'
}>()

const orderQuery = useQuery({
  queryKey: computed(() => [
    'order-list',
    props.orderID,
    props.target,
    page.value,
  ]),
  queryFn: async () =>
    api.get<OwnedOrderItemResponse[]>(
      `orders/items/${props.target}/${props.orderID}`,
      {
        headers: {
          Authorization: store.getAccessToken,
        },
        params: {
          page: page.value,
        },
      },
    ),
  select(res) {
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
    return res.data
  },
})

watch(orderQuery.error, (err) => {
  if (err !== undefined) {
    errorStatusHandler(err, router, {
      notFound() {
        // empty block
        return
      },
    })
  }
})
</script>

<template>
  <section class="order-list-view c-flex-all-center">
    <div
      v-if="
        orderQuery.data.value === undefined ||
        orderQuery.data.value.length === 0
      "
      class="w-ful p-4 text-2xl font-bold"
    >
      <h2>Empty Order</h2>
    </div>
    <div class="c-flex-all-center flex-col w-full gap-5">
      <div class="c-flex-all-center flex-col w-full gap-5">
        <div
          v-for="item of orderQuery.data.value"
          :key="item.product_id"
          class="item"
        >
          <h3>Name: {{ item.name }}</h3>
          <p>Product ID: {{ item.product_id }}</p>
          <p>Item Total: {{ item.items_total }}</p>
          <p>Processed Items: {{ item.processed_items }}</p>
          <div class="c-flex-all-center flex-col gap-5 w-full">
            <h3>Vendor List:</h3>
            <ul
              v-for="vendors of item.confirmed_vendors"
              :key="vendors.user_id"
            >
              <li>Shop ID: {{ vendors.user_id }}</li>
              <li>Quantity: {{ vendors.quantity }}</li>
            </ul>
          </div>
        </div>
      </div>
      <div class="w-full">
        <ListPagination
          :last="maxPage"
          :page-name="`order-list-${props.target}`"
          @change-page="
            (p) => {
              page = p
            }
          "
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.order-list-view {
  @apply w-full h-fit flex-col mt-8;
}

.order-list-view .item {
  @apply shadow hover:shadow-[0_0_10px_5px] shadow-(color:--c-v-10)
    w-95/100 sm:w-80/100 flex items-start justify-evenly flex-col
    min-h-60 p-4 rounded-2xl;
}

.order-list-view h3 {
  @apply text-2xl font-bold p-4 border-b border-(--c-v-10)
    w-full mb-4;
}

.order-list-view p {
  @apply text-wrap wrap-break-word w-full;
}

.order-list-view ul {
  @apply rounded-2xl bg-(--c-v-5) dark:bg-(--c-v-3)
    shadow p-2 w-full;
}

.order-list-view li {
  @apply w-full text-wrap wrap-break-word;
}
</style>
