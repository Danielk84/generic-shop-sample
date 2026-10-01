<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { PermissionString } from '@/contracts/users/request.schema'
import type { UserResponse } from '@/contracts/users/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const page = ref<number>(1)
const maxPage = ref<number>(1)

const store = useStore()
const router = useRouter()

const { data, error } = useQuery({
  queryKey: computed(() => ['admin-users-list', page.value]),
  queryFn: async () =>
    api.get<UserResponse[]>('users/', {
      params: {
        page: page.value,
      },
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
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
      errorStatusHandler(err, router)
    }
  },
  {
    immediate: true,
  },
)
window.scroll(0, 0)
</script>

<template>
  <div class="users-list-page">
    <div v-if="data === undefined || data.length === 0" class="empty">
      <p>No users found.</p>
    </div>
    <div v-else class="list">
      <RouterLink
        v-for="item in data"
        :key="item.id"
        class="item"
        :to="{ name: 'admin-user-info', params: { id: item.id } }"
      >
        <div class="row">
          <span class="label">Name</span>
          <span>{{
            item.name === '' || item.name === ' ' || item.name === undefined
              ? 'Unknown!?'
              : item.name
          }}</span>
        </div>
        <div class="row">
          <span class="label">Permission</span>
          <span>{{ PermissionString(item.permission_type) }}</span>
        </div>
        <div class="badges">
          <span class="badge" :class="item.is_active ? 'ok' : 'bad'">
            {{ item.is_active ? 'Active' : 'Inactive' }}
          </span>
          <span class="badge" :class="item.is_verified ? 'ok' : 'bad'">
            {{ item.is_verified ? 'Verified' : 'Unverified' }}
          </span>
        </div>
      </RouterLink>
    </div>
    <ListPagination
      :last="maxPage"
      page-name="admin-users-list"
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

.users-list-page {
  @apply w-full min-h-screen p-10
    flex flex-col items-center gap-8;
}

.empty {
  @apply text-2xl font-bold p-10;
}

.list {
  @apply w-full max-w-200 flex flex-col gap-4;
}

.item {
  @apply w-full p-5 rounded-2xl
    border border-(--c-v-5)
    flex flex-col gap-4
    hover:shadow-[0_0_15px_4px_var(--c-v-9)]
    transition duration-150;
}

.row {
  @apply flex flex-row justify-between
    border-b border-(--c-v-8);
}

.label {
  @apply font-bold text-(--c-v-4);
}

.badges {
  @apply flex flex-row gap-2;
}

.badge {
  @apply px-3 py-1 rounded-xl text-sm font-bold;
}

.badge.ok {
  @apply bg-(--c-v-14) text-(--c-v-1);
}

.badge.bad {
  @apply bg-(--c-v-11) text-(--c-v-1);
}
</style>
