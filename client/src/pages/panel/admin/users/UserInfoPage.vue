<script setup lang="ts">
import { defineAsyncComponent, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { PermissionString } from '@/contracts/users/request.schema'
import type { UserDetailResponse } from '@/contracts/users/response.interface'

const UpdatePermission = defineAsyncComponent(
  () => import('@/pages/panel/admin/users/UpdatePermission.vue'),
)

const store = useStore()
const route = useRoute()
const router = useRouter()
const id = route.params.id

const { data, error } = useQuery({
  queryKey: ['admin-user-info', id],
  queryFn: async () =>
    api.get<UserDetailResponse>(`users/${id}`, {
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
  <div class="user-info-page">
    <div v-if="data !== undefined" class="card">
      <p>ID: {{ data.id }}</p>
      <h1>Name: {{ data.name }}</h1>
      <h3>Verified: {{ data.is_verified }}</h3>
      <h2>Email: {{ data.email }}</h2>
      <h3>Verified email: {{ data.is_v_email }}</h3>
      <h2>Phone number: {{ data.phone_number }}</h2>
      <h3>Verified phone number: {{ data.is_v_phone_number }}</h3>
      <h2>Permission type: {{ PermissionString(data.permission_type) }}</h2>
    </div>
    <UpdatePermission />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.user-info-page {
  @apply w-full min-h-150 p-10
    flex flex-col items-center gap-8;
}

.card {
  @apply w-full max-w-200 p-6 rounded-2xl
    border border-(--c-v-5)
    flex flex-col gap-2;
}
</style>
