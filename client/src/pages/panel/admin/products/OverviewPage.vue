<script setup lang="ts">
import { defineAsyncComponent, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { ProductResponse } from '@/contracts/products/response.interface'

const ProductView = defineAsyncComponent(
  () => import('@/components/common/products/ProductView.vue'),
)

const store = useStore()
const route = useRoute()
const router = useRouter()
const id = route.params.productID

const { data, error } = useQuery({
  queryKey: ['admin-product-info', id],
  queryFn: () =>
    api.get<ProductResponse>(`products/overview/${id}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  select: (res) => res.data,
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router)
    }
  },
  {
    immediate: true,
  },
)
</script>

<template>
  <div class="w-screen min-h-screen">
    <ProductView :id="id as string" :data="data" />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
