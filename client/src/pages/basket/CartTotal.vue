<script setup lang="ts">
import type { CartToralProps } from '@/pages/basket/types'

const props = defineProps<{ data: CartToralProps }>()

function dashIfUndefined(v?: number) {
  if (v) return v
  return '-'
} 
</script>

<template>
  <div class="card-total">
    <div class="item title">
      <h2>Cart Total</h2>
    </div>
    <table>
      <tr class="item">
        <th>Total Items</th>
        <td>{{ props.data.totalItem }}</td>
      </tr>
      <tr class="item">
        <th>Discount</th>
        <td>{{ dashIfUndefined(props.data.discount) }}</td>
      </tr>
      <tr class="item">
        <th>Total</th>
        <td>{{ props.data.total }}</td>
      </tr>
    </table>
    <div class="item">
      <button class="c-flex-all-center">
        <span>Process Order</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.card-total {
  @apply w-fit h-fit rounded-xl
    border border-basket-card-total-border
    overflow-hidden;
}

.card-total .item {
  @apply text-2xl flex flex-row
    items-center h-20 w-150;
}

.card-total .title {
  @apply justify-start p-4
    bg-basket-card-total-title
    text-basket-card-total-title-text;
}

.card-total tr {
  @apply justify-between border-b p-4
    border-basket-card-total-border;
}

.card-total button {
  @apply size-full cursor-pointer
    bg-basket-card-total-btn
    text-basket-card-total-btn-text;
}
</style>