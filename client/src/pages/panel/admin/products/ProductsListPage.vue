<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler, formatDate } from '@/utils/helper'
import type { ProductStatusResponse } from '@/contracts/products/response.interface'

const ImageFrameCard = defineAsyncComponent(
  () => import('@/components/card/ImageFrameCard.vue'),
)
const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)

const store = useStore()
const route = useRoute()
const router = useRouter()

const page = ref<number | null>(null)
const maxPage = ref<number>(1)

const { data, error, refetch } = useQuery({
  queryKey: ['admin-products-list', page.value],
  enabled: computed(() => page.value !== null),
  queryFn: async () =>
    api.get<ProductStatusResponse[]>('products/admin', {
      params: {
        page: page.value,
      },
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  select: (res) => {
    const mp = Number(res.headers['x-max-page'])
    if (mp !== undefined) {
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
          const p = route.query?.page
          if (p !== null && Number(p) > 1) {
            router.back()
          }
        },
      })
    }
  },
  {
    immediate: true,
  },
)

const setActiveMutation = useMutation({
  mutationKey: ['admin-products-set-active'],
  mutationFn: async (input: { id: string; status: boolean }) => {
    return api.put(
      `products/set-active/${input.id}`,
      { accepted: input.status },
      {
        headers: {
          Authorization: store.getAccessToken,
        },
      },
    )
  },
  onSuccess: async () => {
    await refetch()
  },
  onError: (error) => {
    errorStatusHandler(error, router)
  },
})
</script>

<template>
  <div class="products-list-page">
    <div class="options">
      <RouterLink
        class="redirect-btn"
        :to="{
          name: 'admin-product-create',
          query: { page: page },
        }"
      >
        Create
      </RouterLink>
    </div>
    <div v-if="data === undefined" class="empty-product c-flex-all-center">
      <p>There are not any products.</p>
    </div>
    <div
      v-else
      v-for="item in data"
      :key="item.id"
      class="list c-flex-all-center"
    >
      <div class="item">
        <div class="img-frame">
          <ImageFrameCard :img="item.img_path" loading="eager" />
        </div>
        <div class="content">
          <h2>{{ item.name }}</h2>
          <p>Price: {{ item.price }}</p>
          <p>Published: {{ formatDate(item.pub_date) }}</p>
          <p>Quantity: {{ item.available_quantity }}</p>
          <button
            class="set-btn true-btn cursor-pointer"
            :class="{
              'false-btn': !item.is_active,
              'c-is-pending': !setActiveMutation.isPending,
            }"
            @click="
              setActiveMutation.mutate({
                id: item.id,
                status: !item.is_active,
              })
            "
          >
            Active
          </button>
          <div
            class="set-btn true-btn"
            :class="{ 'false-btn': !item.is_available }"
          >
            <span>Available</span>
          </div>
          <RouterLink
            class="set-btn base-btn"
            :to="{
              name: 'admin-product-edit',
              params: { productID: item.id },
              query: { page: page },
            }"
          >
            <span>Edit</span>
          </RouterLink>
          <RouterLink class="set-btn base-btn" :to="`products/${item.id}`">
            <span>Overeview</span>
          </RouterLink>
        </div>
      </div>
    </div>
    <div class="pagination">
      <ListPagination
        :last="maxPage"
        page-name="admin-products-list"
        @change-page="
          (p: number) => {
            page = p
          }
        "
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.products-list-page {
  @apply w-screen min-h-full flex flex-col items-center;
}

.products-list-page .options {
  @apply p-10 w-full border-b border-panel-products-border
    flex flex-row items-center justify-center
    text-2xl font-bold;
}

.products-list-page .redirect-btn {
  @apply bg-panel-products-create-btn rounded-2xl px-10 py-5
    hover:brightness-110;
}

.products-list-page .empty-product {
  @apply text-2xl font-bold p-10;
}

.products-list-page .list {
  @apply flex-col gap-10 p-10 size-full;
}

.products-list-page .item {
  @apply w-223 h-97 p-8
    rounded-2xl shadow hover:shadow-[0_0_20px_5px]
    shadow-panel-products-shadow
    flex flex-row items-center justify-evenly;
}

.products-list-page .img-frame {
  @apply size-82 rounded-2xl overflow-hidden
    basis-3/7;
}

.products-list-page .content {
  @apply flex flex-col gap-2
    border-l px-4
    border-panel-products-shadow
    basis-4/7;
}

.content h2 {
  @apply font-bold;
}

.products-list-page .set-btn {
  @apply w-40 h-10 rounded-2xl
    flex items-center justify-center
    text-xl font-bold
    hover:brightness-120;
}

.products-list-page .true-btn {
  @apply bg-panel-products-ture-btn
    text-panel-products-btn-text;
}

.products-list-page .false-btn {
  @apply bg-panel-products-false-btn;
}

.products-list-page .base-btn {
  @apply border-2 border-panel-products-border
    cursor-pointer;
}
</style>
