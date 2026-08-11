<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

 import type { OrderItemsCardProps } from '@/pages/basket/types'


const OrderItemsCard = defineAsyncComponent(() => import('@/pages/basket/OrderItemsCard.vue'))

const props = defineProps<{ orders: OrderItemsCardProps[] }>()
</script>

<template>
  <div class="order-items">
    <div class="title c-flex-all-center">
      <span>Products</span>
    </div>
    <div
      v-if="props.orders.length === 0"
      class="empty c-flex-all-center"
    >
      <span>empty!</span>
    </div>
    <div v-else class="orders">
      <OrderItemsCard v-for="i of props.orders" :key="i.name" :data="{
        name: i.name,
        price: i.price,
        count: i.count,
        total: i.total,
      }" />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.order-items {
  @apply w-150 h-fit
    border rounded-t-2xl overflow-hidden
    border-basket-orders-border;
}

.order-items > .title {
  @apply text-2xl font-black w-full h-20
    bg-basket-orders-title-bg
    text-basket-orders-title-text;
}

.order-items > .orders {
  @apply w-fit h-fit;
}

.order-items > .empty {
  @apply text-2xl py-10 font-bold;
}
</style>