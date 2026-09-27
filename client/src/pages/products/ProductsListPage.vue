<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
import { SearchRequest } from '@/contracts/search/request.schema'
import type { ProductSummaryResponse } from '@/contracts/products/response.interface'

const SearchBox = defineAsyncComponent(
  () => import('@/components/ui/search/SearchBox.vue'),
)
const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)
const ProductCard = defineAsyncComponent(
  () => import('@/components/card/ProductCard.vue'),
)

const route = useRoute()
const router = useRouter()

const page = ref<number>(1)
const maxPage = ref<number>(1)
const isPopular = computed(() => route.name === 'products-popular')
const searchQuery = computed(() => {
  const q = route.query.q
  return typeof q === 'string' && q !== '' ? q : null
})

const { data, error } = useQuery({
  queryKey: computed(() => [
    'products-list',
    page.value,
    searchQuery.value ?? '',
    isPopular.value ? 'popular' : 'all',
  ]),
  queryFn: async () => {
    if (searchQuery.value !== null) {
      const input = SearchRequest.parse({ query_str: searchQuery.value })
      return api.post<ProductSummaryResponse[]>('search/', input, {
        params: { page: page.value },
      })
    }
    if (isPopular.value) {
      return api.get<ProductSummaryResponse[]>('products/popular', {
        params: { page: page.value },
      })
    }
    return api.get<ProductSummaryResponse[]>('products/', {
      params: { page: page.value },
    })
  },
  select: (res) => {
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
    return res.data
  },
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          const p = route.query?.page
          if (p !== null && Number(p) > 1) {
            router.back()
          }
        },
      })
    }
  },
  {
    immediate: true,
  },
)

watch([searchQuery, isPopular], () => {
  page.value = 1
})
</script>

<template>
  <div class="products-list">
    <div class="products-search">
      <SearchBox />
    </div>
    <div v-if="data === undefined || data.length === 0" class="empty">
      <h2>
        <span>Sorry!</span>
        <span>But it is empty!</span>
      </h2>
    </div>
    <div v-else class="items">
      <ProductCard
        v-for="item in data"
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
    <ListPagination
      :last="maxPage"
      :page-name="String(route.name)"
      @change-page="
        (newValue: number) => {
          page = newValue
        }
      "
    />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.products-list {
  @apply w-full min-h-200 p-8
    flex flex-col items-center justify-between gap-10;
}

.products-list > .products-search {
  @apply w-80/100 h-fit;
}

.products-list > .empty > h2 {
  @apply text-4xl flex flex-col items-start justify-center;
}

.products-list > .items {
  @apply w-full flex flex-row flex-wrap
    justify-center items-start gap-10 p-4;
}
</style>
