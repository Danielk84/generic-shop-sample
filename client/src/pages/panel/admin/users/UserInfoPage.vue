<script setup lang="ts">
import { watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { UserDetailResponse } from '@/contracts/users/response.interface'

const store = useStore()
const route = useRoute()
const router = useRouter()
const id = route.params.id

const { data, error } = useQuery({
  queryKey: ['admin-user-info', id],
  queryFn: async () =>
    api.get<UserDetailResponse>(`user/${id}`, {
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
      <p>ID: {{ data.id }}</p>
      <h1>Name: {{ data.name }}</h1>
      <h3>Verified: {{ data.is_verified }}</h3>
      <h2>Email: {{ data.email }}</h2>
      <h3>Verified email: {{ data.is_v_email }}</h3>
      <h2>Phone number: {{ data.phone_number }}</h2>
      <h3>Verified phone number: {{ data.is_v_phone_number }}</h3>
      <h2>Permission type: {{ data.permission_type }}</h2>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
