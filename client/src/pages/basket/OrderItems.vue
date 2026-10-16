<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import type { OrderItemsCardProps } from '@/pages/basket/types'

const OrderItemsCard = defineAsyncComponent(
  () => import('@/pages/basket/OrderItemsCard.vue'),
)

const props = defineProps<{
  orders: (OrderItemsCardProps & { product_id: string })[]
}>()

const emits = defineEmits<{
  (e: 'remove', productId: string): void
  (e: 'changeCount', input: { productId: string; count: number }): void
}>()
</script>

<template>
  <div class="order-items">
    <div class="title c-flex-all-center">
      <span>Products</span>
    </div>
    <div v-if="props.orders.length === 0" class="empty c-flex-all-center">
      <span>empty!</span>
    </div>
    <div v-else class="orders">
      <OrderItemsCard
        v-for="i of props.orders"
        :key="i.product_id"
        :data="{
          name: i.name,
          price: i.price,
          count: i.count,
          total: i.total,
          img: i.img,
          alt: i.alt,
        }"
        @destroy="emits('remove', i.product_id)"
        @change-count="
          (count: number) =>
            emits('changeCount', { productId: i.product_id, count })
        "
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.order-items {
  @apply w-full max-w-200 h-fit
    border rounded-t-2xl overflow-hidden
    border-(--c-v-1);
}

.order-items > .title {
  @apply text-xl md:text-2xl font-black w-full h-20
    bg-(--c-v-1)
    text-(--c-v-7);
}

.order-items > .orders {
  @apply w-full h-fit;
}

.order-items > .empty {
  @apply text-xl md:text-2xl py-10 font-bold;
}
</style>
