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

const router = useRouter()

const { data, isPending, error } = useQuery({
  queryKey: ['latest-products'],
  queryFn: async () => api.get<ProductSummaryResponse[]>('products/'),
  select: (res) => res.data,
})

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
    title="New Products"
    to="/products/"
    empty-view="Sorry, but there is not any new products."
    :data="data"
    :loading="isPending"
  />
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
