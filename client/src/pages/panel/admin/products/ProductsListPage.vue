<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler, formatDate } from '@/utils/helper'
import { SearchRequest } from '@/contracts/search/request.schema'
import type {
  ProductStatusResponse,
  ProductSummaryResponse,
} from '@/contracts/products/response.interface'

const ImageFrameCard = defineAsyncComponent(
  () => import('@/components/card/ImageFrameCard.vue'),
)
const ListPagination = defineAsyncComponent(
  () => import('@/components/ui/ListPagination.vue'),
)
const DeleteProduct = defineAsyncComponent(
  () => import('@/pages/panel/admin/products/DeleteProduct.vue'),
)

const store = useStore()
const route = useRoute()
const router = useRouter()
const notification = useNotificationStore()

const page = ref<number | null>(null)
const maxPage = ref<number>(1)
const searchText = ref('')
const searchResults = ref<ProductSummaryResponse[] | null>(null)

const { data, error, refetch } = useQuery({
  queryKey: computed(() => ['admin-products-list', page.value]),
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

const searchMutation = useMutation({
  mutationKey: ['admin-products-search-all'],
  mutationFn: async (input: { q: string; page: number }) => {
    const parsed = SearchRequest.parse({ query_str: input.q })
    return api.post<ProductSummaryResponse[]>('search/all', parsed, {
      params: { page: input.page },
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: (res) => {
    searchResults.value = res.data
    const mp = Number(res.headers['x-max-page'])
    if (!Number.isNaN(mp) && mp > 0) {
      maxPage.value = mp
    }
  },
  onError: (err) => {
    searchResults.value = null
    errorStatusHandler(err, router, {
      notFound() {
        // empty back
        return
      },
      badRequest() {
        notification.warning('invalid search input')
        return
      },
    })
  },
})

const reindexMutation = useMutation({
  mutationKey: ['admin-products-reindex'],
  mutationFn: async (id: string) => {
    return api.get(`search/reindex/${id}`, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: () => {
    notification.success('Product queued for reindex.')
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
})

function runSearch() {
  const q = searchText.value.trim()
  if (q === '') return
  searchMutation.mutate({ q, page: page.value ?? 1 })
}

function clearSearch() {
  searchText.value = ''
  searchResults.value = null
}

watch(page, (v) => {
  const q = searchText.value.trim()
  if (q !== '' && searchResults.value !== null && v !== null) {
    searchMutation.mutate({ q, page: v })
  }
})
</script>

<template>
  <div class="products-list-page">
    <section class="options">
      <RouterLink
        class="redirect-btn"
        :to="{
          name: 'admin-product-create',
          query: { page: page },
        }"
      >
        Create
      </RouterLink>
      <form
        class="c-form c-form-bg p-5 rounded-2xl"
        role="search"
        @submit.prevent="runSearch"
      >
        <input
          class="c-form-input"
          id="admin-product-search"
          v-model="searchText"
          type="search"
          placeholder="Full Search"
          autocomplete="off"
        />
        <button
          class="c-form-btn"
          type="submit"
          :disabled="searchMutation.isPending.value"
        >
          Search
        </button>
        <button
          v-if="searchResults !== null"
          type="button"
          class="c-form-btn m-0"
          @click="clearSearch"
        >
          Clear
        </button>
      </form>
    </section>
    <section
      v-if="searchResults !== null"
      class="search-results c-flex-all-center"
    >
      <div
        v-if="searchResults.length === 0"
        class="empty-product c-flex-all-center"
      >
        <p>No products match your search.</p>
      </div>
      <div
        v-else
        v-for="item in searchResults"
        :key="item.id"
        class="list c-flex-all-center"
      >
        <div class="item">
          <div class="img-frame">
            <ImageFrameCard :img="item.img_path" loading="lazy" />
          </div>
          <div class="content">
            <h2>{{ item.name }}</h2>
            <p>Price: {{ item.price }}</p>
            <p>Published: {{ formatDate(item.pub_date) }}</p>
            <RouterLink
              class="set-btn base-btn"
              :to="{
                name: 'admin-product-overview',
                params: { productID: item.id },
              }"
            >
              <span>Overview</span>
            </RouterLink>
            <button
              type="button"
              class="set-btn base-btn cursor-pointer"
              :disabled="reindexMutation.isPending.value"
              @click="reindexMutation.mutate(item.id)"
            >
              Reindex
            </button>
          </div>
        </div>
      </div>
    </section>
    <div v-if="data === undefined" class="empty-product c-flex-all-center">
      <p>There are not any products.</p>
    </div>
    <div
      v-else-if="searchResults === null"
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
          <RouterLink
            class="set-btn base-btn"
            :to="{
              name: 'admin-product-overview',
              params: {
                productID: item.id,
              },
            }"
          >
            <span>Overview</span>
          </RouterLink>
          <button
            type="button"
            class="set-btn base-btn cursor-pointer"
            :disabled="reindexMutation.isPending.value"
            @click="reindexMutation.mutate(item.id)"
          >
            Reindex
          </button>
          <DeleteProduct
            class="set-btn base-btn bg-(--c-v-11)"
            :id="item.id"
            @success="refetch"
          />
        </div>
      </div>
    </div>
    <div class="p-10">
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
  @apply w-screen min-h-screen flex flex-col items-center;
}

.products-list-page .options {
  @apply p-10 w-full border-b border-(--c-v-1)
    flex flex-col items-center justify-center gap-10
    text-2xl font-bold;
}

.products-list-page .redirect-btn {
  @apply bg-(--c-v-15) rounded-2xl px-10 py-5
    hover:brightness-110;
}

.products-list-page .empty-product {
  @apply text-2xl font-bold p-10;
}

.products-list-page .list {
  @apply flex-col gap-10 p-10 size-full;
}

.products-list-page .item {
  @apply w-223 min-h-97 p-8
    rounded-2xl
    shadow-[0_1px_3px_0_var(--c-v-9),0_1px_2px_-1px_var(--c-v-9)]
    hover:shadow-[0_0_20px_5px_var(--c-v-9)]
    flex flex-row items-center justify-evenly;
}

.products-list-page .img-frame {
  @apply size-82 rounded-2xl overflow-hidden
    basis-3/7;
}

.products-list-page .content {
  @apply flex flex-col gap-2
    border-l px-4
    border-(--c-v-9)
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
  @apply bg-(--c-v-14)
    text-(--c-v-1);
}

.products-list-page .false-btn {
  @apply bg-(--c-v-11) text-(--c-v-7);
}

.products-list-page .base-btn {
  @apply border-2 border-(--c-v-1)
    cursor-pointer;
}
</style>
