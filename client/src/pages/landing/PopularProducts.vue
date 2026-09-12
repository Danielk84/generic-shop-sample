<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import type { ProductSummaryResponse } from '@/contracts/products/response.interface'

const ProductsView = defineAsyncComponent(
  () => import('@/pages/landing/ProductsView.vue'),
)

const data: ProductSummaryResponse[] = []
</script>

<template>
  <ProductsView
    title="New Products"
    to="/"
    empty-view="Sorry, but there is not popular products."
    :data="data"
  />
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
