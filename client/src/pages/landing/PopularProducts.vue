<script setup lang="ts">
import { defineAsyncComponent, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
import type { ProductSummaryResponse } from '@/contracts/products/response.interface'

const ProductsView = defineAsyncComponent(
  () => import('@/pages/landing/ProductsView.vue'),
)

const emits = defineEmits<{
  (e: 'setLoadingState', state: boolean): void
}>()

const router = useRouter()

const { data, isPending, isFetched, error } = useQuery({
  queryKey: ['papular-products'],
  queryFn: async () => api.get<ProductSummaryResponse[]>('/products/popular'),
  select: (res) => res.data,
})

watch(
  () => ({
    isPending: isPending.value,
    isFetched: isFetched.value,
  }),
  (state) => {
    emits('setLoadingState', state.isFetched && state.isPending)
  },
  {
    immediate: true,
  },
)

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          // do nothing.
          return
        },
      })
    }
  },
  {
    immediate: true,
  },
)
</script>

<template>
  <ProductsView
    title="Popular Products"
    to="/"
    empty-view="Sorry, but there is not any popular products."
    :data="data"
  />
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
