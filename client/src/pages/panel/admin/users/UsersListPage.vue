<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import * as products_res from '@/contracts/users/response.interface'

const page = ref<number>(1)

const store = useStore()
const router = useRouter()

const { data, error } = useQuery({
  queryKey: ['admin-users-list'],
  queryFn: async () =>
    api.get<products_res.UserResponse[]>('users/', {
      params: {
        page,
      },
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
    <RouterLink
      v-for="item in data"
      :key="item.id"
      :to="{ name: 'admin-user-info', params: { id: item.id } }"
    >
      <h2>{{ item.name }}</h2>
      <h3>{{ item.permission_type }}</h3>
      <p>Active: {{ item.is_active }}</p>
      <p>Verified: {{ item.is_verified }}</p>
    </RouterLink>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
