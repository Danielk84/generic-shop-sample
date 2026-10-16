<script setup lang="ts">
import type { CartToralProps } from '@/pages/basket/types'

const props = withDefaults(
  defineProps<{
    data: CartToralProps
    disabled?: boolean
    processing?: boolean
  }>(),
  {
    disabled: false,
    processing: false,
  },
)
const emits = defineEmits<{
  (e: 'process'): void
}>()

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
    <table class="w-full">
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
      <button
        class="c-flex-all-center"
        :disabled="props.disabled || props.processing"
        @click="emits('process')"
      >
        <span>{{ props.processing ? 'Processing...' : 'Process Order' }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.card-total {
  @apply w-full max-w-100 rounded-xl
    border border-(--c-v-1)
    overflow-hidden;
}

.card-total .item {
  @apply text-lg md:text-2xl flex flex-row
    items-center min-h-20;
}

.card-total .title {
  @apply justify-start p-4
    bg-(--c-v-4)
    text-(--c-v-7);
}

.card-total tr {
  @apply justify-between border-b p-4 w-full
    border-(--c-v-1);
}

.card-total button {
  @apply w-full h-12 cursor-pointer
    bg-(--c-v-1)
    text-(--c-v-7);
}

.card-total button:disabled {
  @apply opacity-50 cursor-not-allowed;
}
</style>
