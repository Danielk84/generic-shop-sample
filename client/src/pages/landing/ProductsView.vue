<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import type { ProductSummaryResponse } from '@/contracts/products/response.interface'

const ProductCard = defineAsyncComponent(
  () => import('@/components/card/ProductCard.vue'),
)

const props = defineProps<{
  title: string
  to: string
  emptyView: string
  data: ProductSummaryResponse[]
}>()
</script>

<template>
  <div v-if="props.data.length === 0" class="empty-view c-flex-all-center">
    <h2>{{ props.emptyView }}</h2>
  </div>
  <div v-else class="products-view">
    <div class="label c-flex-all-center">
      <h2>New Products</h2>
      <div class="line"></div>
      <RouterLink class="link" to="/"> See more... </RouterLink>
    </div>
    <div class="products">
      <ProductCard
        class="product-item"
        v-for="item of props.data"
        :key="item.id"
        :data="{
          to: `/products/${item.id}`,
          title: item.name,
          price: item.price,
          img: item.img_path,
          alt: `${item.name}-${item.pub_date}`,
        }"
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.empty-view {
  @apply font-bold text-4xl w-full h-fit p-20;
}

.products-view {
  @apply w-full h-fit px-10;
}

.products-view .label {
  @apply gap-5;
}

.label > h2 {
  @apply font-bold text-4xl text-nowrap;
}

.label > .line {
  @apply w-full h-fit border-2
    border-products-view-line;
}

.label > .link {
  @apply text-2xl font-bold text-nowrap
    text-products-view-link;
}

.products-view .products {
  @apply flex flex-row items-start
    gap-20 p-10 w-full h-fit
    overflow-x-auto flex-nowrap
    snap-x scroll-pl-20 scroll-smooth;
}

.products .product-item {
  @apply shrink-0 snap-start;
}
</style>
