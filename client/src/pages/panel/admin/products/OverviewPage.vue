<script setup lang="ts">
import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { ProductResponse } from '@/contracts/products/response.interface'

const store = useStore()
const route = useRoute()
const router = useRouter()
const id = route.params.id

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
  <div>
    <div v-if="data !== undefined">
      <p>{{ data.id }}</p>
      <p>{{ data.name }}</p>
      <p>{{ data.price }}</p>
      <p>{{ data.pub_date }}</p>
      <p>{{ data.available_quantity }}</p>
      <p>{{ data.is_available }}</p>
      <p>{{ data.is_active }}</p>
      <p>{{ data.description }}</p>
      <p>{{ data.common_detail }}</p>
      <p>{{ data.variant_detail }}</p>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
