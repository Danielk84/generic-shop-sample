<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { ShopResponse } from '@/contracts/users/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const store = useStore()
const route = useRoute()
const router = useRouter()

function initialPage(): number {
  const queryPage = Number(route.query.page)
  return Number.isInteger(queryPage) && queryPage > 0 ? queryPage : 1
}

const page = ref<number>(initialPage())
const maxPage = ref<number>(1)

const { data, error, isPending } = useQuery({
  queryKey: computed(() => ['admin-shops-list', page.value]),
  queryFn: async () =>
    api.get<ShopResponse[]>('users/shop/', {
      params: { page: page.value },
      headers: { Authorization: store.getAccessToken },
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
  { immediate: true },
)
</script>

<template>
  <section class="shops-list-page">
    <h1 class="page-title">Shops Management</h1>

    <div v-if="isPending" class="empty">
      <p>Loading shops...</p>
    </div>
    <div v-else-if="data === undefined || data.length === 0" class="empty">
      <p>No shops found.</p>
    </div>
    <div v-else class="list">
      <div v-for="item in data" :key="item.user_id" class="item">
        <div class="row">
          <span class="label">User ID</span>
          <span>{{ item.user_id }}</span>
        </div>
        <div class="row">
          <span class="label">Brand</span>
          <span>{{ item.brand }}</span>
        </div>
        <div class="row">
          <span class="label">Phone</span>
          <span>{{ item.phone_number }}</span>
        </div>
        <div class="badges">
          <span class="badge" :class="item.is_shop ? 'ok' : 'bad'">
            {{ item.is_shop ? 'Active' : 'Inactive' }}
          </span>
        </div>
        <div class="actions">
          <RouterLink
            class="btn toggle-btn"
            :to="{ name: 'admin-shops-detail', params: { id: item.user_id } }"
          >
            See Detail
          </RouterLink>
        </div>
      </div>
    </div>

    <ListPagination
      :last="maxPage"
      page-name="admin-shops-list"
      @change-page="
        (newValue: number) => {
          page = newValue
        }
      "
    />
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.shops-list-page {
  @apply w-full min-h-screen p-4 md:p-10
    flex flex-col items-center gap-6 md:gap-8;
}

.page-title {
  @apply text-2xl md:text-3xl font-bold;
}

.empty {
  @apply text-xl md:text-2xl font-bold p-6 md:p-10;
}

.list {
  @apply w-full max-w-200 flex flex-col gap-4;
}

.item {
  @apply w-full p-4 md:p-5 rounded-2xl
    border border-(--c-v-5)
    flex flex-col gap-3 md:gap-4
    hover:shadow-[0_0_15px_4px_var(--c-v-9)]
    transition duration-150;
}

.row {
  @apply flex flex-row justify-between gap-2 flex-wrap
    border-b border-(--c-v-8) pb-2;
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
  @apply bg-(--c-v-11) text-(--c-v-7);
}

.actions {
  @apply flex flex-col md:flex-row gap-2 md:gap-4 mt-2;
}

.btn {
  @apply w-full md:w-50 h-10 md:h-12 rounded-xl font-bold text-base md:text-lg cursor-pointer
    hover:brightness-95 flex items-center justify-center;
}

.toggle-btn {
  @apply bg-(--c-v-8) text-(--c-v-1);
}
</style>
