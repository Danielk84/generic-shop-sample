<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useNotificationStore } from '@/store/notification'
import type { VendorOrder } from '@/contracts/vendors/response.interface'

const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const route = useRoute()
const router = useRouter()
const store = useStore()
const notification = useNotificationStore()

const page = ref<number>(1)
const maxPage = ref<number>(1)

const { data, error, refetch } = useQuery({
  queryKey: computed(() => ['vendor-orders', page.value]),
  queryFn: async () =>
    api.get<VendorOrder[]>('vendors/', {
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
      errorStatusHandler(err, router, {
        notFound() {
          const p = Number(route.query?.page)
          if (p > 1) router.back()
        },
      })
    }
  },
  { immediate: true },
)

const setDeliveredMutation = useMutation({
  mutationKey: ['vendor-set-delivered'],
  mutationFn: async (input: VendorOrder) => {
    return api.put(
      'vendors/set-delivered',
      {
        user_id: input.user_id,
        order_id: input.order_id,
        product_id: input.product_id,
        property: input.property,
        is_delivered: !input.is_delivered,
      },
      { headers: { Authorization: store.getAccessToken } },
    )
  },
  onSuccess: async () => {
    notification.success('Delivery status updated.')
    await refetch()
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
})
</script>

<template>
  <div class="vendor-orders-page">
    <h1 class="page-title">Vendor Orders</h1>
    <div v-if="data === undefined || data.length === 0" class="empty">
      <p>No vendor orders.</p>
    </div>
    <div v-else class="orders">
      <div
        v-for="item in data"
        :key="`${item.order_id}-${item.product_id}-${JSON.stringify(item.property)}`"
        class="order-item"
      >
        <div class="row">
          <span class="label">Order</span>
          <span class="mono">{{ item.order_id }}</span>
        </div>
        <div class="row">
          <span class="label">Product</span>
          <span class="mono">{{ item.product_id }}</span>
        </div>
        <div class="row">
          <span class="label">Quantity</span>
          <span>{{ item.quantity }}</span>
        </div>
        <div class="row">
          <span class="label">Total</span>
          <span>{{ item.total_bill }}</span>
        </div>
        <div class="props">
          <span v-for="(v, k) in item.property" :key="k" class="prop">
            {{ k }}: {{ v }}
          </span>
        </div>
        <div class="actions">
          <span class="badge" :class="item.is_delivered ? 'ok' : 'bad'">
            {{ item.is_delivered ? 'Delivered' : 'Pending' }}
          </span>
          <button
            class="btn"
            :class="{ 'c-is-pending': setDeliveredMutation.isPending.value }"
            :disabled="setDeliveredMutation.isPending.value"
            @click="setDeliveredMutation.mutate(item)"
          >
            {{ item.is_delivered ? 'Mark Undelivered' : 'Mark Delivered' }}
          </button>
        </div>
      </div>
    </div>
    <ListPagination
      :last="maxPage"
      page-name="vendor-orders"
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

.vendor-orders-page {
  @apply w-full min-h-screen p-4 md:p-10
    flex flex-col items-center gap-6 md:gap-8;
}

.page-title {
  @apply text-2xl md:text-3xl font-bold;
}

.empty {
  @apply text-xl md:text-2xl font-bold p-6 md:p-10;
}

.orders {
  @apply w-full max-w-200 flex flex-col gap-4;
}

.order-item {
  @apply w-full p-4 md:p-5 rounded-2xl
    border border-(--c-v-5)
    flex flex-col gap-3;
}

.row {
  @apply flex flex-row justify-between gap-2
    flex-wrap;
}

.label {
  @apply font-bold text-(--c-v-4) whitespace-nowrap;
}

.mono {
  @apply break-all text-right flex-1 text-center md:text-right;
}

.props {
  @apply flex flex-row gap-2 flex-wrap;
}

.prop {
  @apply px-2 py-1 rounded-lg text-sm
    bg-(--c-v-6) dark:bg-(--c-v-3) shadow shadow-(color:--c-v-0);
}

.actions {
  @apply flex flex-row justify-between items-center gap-3 mt-2
    flex-wrap;
}

.badge {
  @apply px-3 py-1 rounded-xl text-sm font-bold
    whitespace-nowrap;
}

.badge.ok {
  @apply bg-(--c-v-14) text-(--c-v-1);
}

.badge.bad {
  @apply bg-(--c-v-11) text-(--c-v-1);
}

.btn {
  @apply px-4 py-2 rounded-xl font-bold cursor-pointer
    bg-(--c-v-9) text-(--c-v-1)
    hover:brightness-95
    whitespace-nowrap;
}

@media (max-width: 640px) {
  .order-item {
    @apply p-4 gap-2;
  }

  .row {
    @apply flex-col gap-1;
  }

  .mono {
    @apply text-left;
  }

  .actions {
    @apply flex-col w-full gap-2;
  }

  .btn {
    @apply w-full;
  }

  .badge {
    @apply w-full text-center;
  }
}
</style>
