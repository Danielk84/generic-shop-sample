<script setup lang="ts">
import { computed, defineAsyncComponent, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler } from '@/utils/helper'
import type { ShopInfoResponse } from '@/contracts/users/response.interface'

const ImageFrameCard = defineAsyncComponent(
  () => import('@/components/card/ImageFrameCard.vue'),
)

const route = useRoute()
const router = useRouter()
const store = useStore()
const notification = useNotificationStore()

const shopId = computed(() => (route.params.id as string) ?? '')

const { data, error, isPending, refetch } = useQuery({
  queryKey: computed(() => ['shop-detail', shopId.value]),
  queryFn: async () =>
    api.get<ShopInfoResponse>(`users/shop/${shopId.value}`, {
      headers: { Authorization: store.getAccessToken },
    }),
  select: (res) => res.data,
  enabled: computed(() => shopId.value !== ''),
})

const setShopStatusMutation = useMutation({
  mutationKey: ['admin-shop-set-active'],
  mutationFn: async (isActive: boolean) =>
    api.put(
      `users/shop/set-active/${shopId.value}`,
      { accepted: isActive },
      { headers: { Authorization: store.getAccessToken } },
    ),
  async onSuccess() {
    notification.success('Shop status updated.')
    await refetch()
  },
  onError(err) {
    errorStatusHandler(err, router)
  },
})

function toggleShopStatus() {
  const isActive = data.value?.is_shop ?? false
  setShopStatusMutation.mutate(!isActive)
}

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
  <div v-if="data !== undefined" class="shop-detail-page">
    <h1 class="page-title">
      {{ data?.brand }}
      <span v-if="data?.is_shop" class="verified-badge">Active</span>
    </h1>

    <div class="card">
      <h2>Shop Information</h2>
      <div class="row">
        <span class="label">Business Code</span>
        <span>{{ data?.business_code }}</span>
      </div>
      <div class="row">
        <span class="label">Address</span>
        <span>{{ data?.shop_addr }}</span>
      </div>
      <div class="row">
        <span class="label">Zip Code</span>
        <span>{{ data?.zip_code }}</span>
      </div>
      <div class="row">
        <span class="label">Phone</span>
        <span>{{ data?.shop_phone_number || 'Not set' }}</span>
      </div>
      <div class="row">
        <span class="label">Bio</span>
        <span>{{ data?.bio }}</span>
      </div>
    </div>

    <div class="card">
      <h2>Shop Status</h2>
      <div class="row">
        <span class="label">Current status</span>
        <span>{{ data?.is_shop ? 'Active' : 'Pending approval' }}</span>
      </div>
      <button
        class="btn primary"
        type="button"
        :disabled="setShopStatusMutation.isPending.value"
        @click="toggleShopStatus"
      >
        {{
          setShopStatusMutation.isPending.value
            ? 'Updating...'
            : data?.is_shop
              ? 'Deactivate shop'
              : 'Activate shop'
        }}
      </button>
    </div>

    <div class="card">
      <h2>Owner Information</h2>
      <div class="row">
        <span class="label">Name</span>
        <span>{{ data?.name }}</span>
      </div>
      <div class="row">
        <span class="label">Email</span>
        <span>{{ data?.email }}</span>
      </div>
      <div class="row">
        <span class="label">Email Verified</span>
        <span>{{ data?.is_v_email ? 'Yes' : 'No' }}</span>
      </div>
      <div class="row">
        <span class="label">Phone</span>
        <span>{{ data?.phone_number }}</span>
      </div>
      <div class="row">
        <span class="label">Phone Verified</span>
        <span>{{ data?.is_v_phone_number ? 'Yes' : 'No' }}</span>
      </div>
    </div>

    <div
      v-if="data?.img_path !== undefined && data?.img_path !== ''"
      class="card c-flex-all-center"
    >
      <h2 class="w-full">Shop Image</h2>
      <div class="size-50 sm:size-100">
        <ImageFrameCard :img="data?.img_path" :alt="data?.brand" />
      </div>
    </div>
  </div>
  <div v-else-if="isPending" class="shop-detail-page empty c-flex-all-center">
    <p>Loading...</p>
  </div>
  <div v-else class="shop-detail-page empty c-flex-all-center">
    <p>Shop not found.</p>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.shop-detail-page {
  @apply w-full min-h-screen p-4 md:p-10
    flex flex-col items-center gap-6 md:gap-8;
}

.page-title {
  @apply text-2xl md:text-3xl font-bold flex flex-row items-center gap-2;
}

.verified-badge {
  @apply px-2 py-1 rounded-xl text-sm font-bold
    bg-(--c-v-14) text-(--c-v-1);
}

.card {
  @apply w-full max-w-200 p-4 md:p-6 rounded-2xl
    border border-(--c-v-5)
    flex flex-col gap-3;
}

.card h2 {
  @apply text-xl md:text-2xl font-bold pb-2 border-b border-(--c-v-4);
}

.row {
  @apply flex flex-col md:flex-row justify-between gap-2;
}

.label {
  @apply font-bold text-(--c-v-4) md:w-1/3;
}

.field {
  @apply flex flex-col gap-1;
}

.field label {
  @apply text-base md:text-lg font-light;
}

.btn {
  @apply h-12 rounded-xl font-bold text-base md:text-lg cursor-pointer
    hover:brightness-95;
}

.btn.primary {
  @apply bg-(--c-v-8) text-(--c-v-7);
}

.btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.form-error {
  @apply text-(--c-v-11) text-sm font-bold;
}

.empty {
  @apply min-h-100 text-xl md:text-2xl font-bold;
}

.form {
  @apply flex flex-col gap-4;
}
</style>
