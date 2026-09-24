<script setup lang="ts">
import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
import type { ProductResponse } from '@/contracts/products/response.interface'
import { isValidID } from '@/utils/validator'

const route = useRoute()
const router = useRouter()
const id = route.params.id

const { error } = useQuery({
  queryKey: ['product-page'],
  queryFn: async () => api.get<ProductResponse>(`products/${id}`),
  select: (res) => res.data,
  enabled: () => isValidID(id),
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
  <div class=""></div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
