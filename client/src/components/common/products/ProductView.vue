<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { defineAsyncComponent, watch } from 'vue'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
import { formatDate } from '@/utils/helper'
import type {
  ProductResponse,
  ProductImageResponse,
  ProductVendor,
} from '@/contracts/products/response.interface'

const ImageFrameList = defineAsyncComponent(
  () => import('@/components/common/products/ImageFrameList.vue'),
)

const props = defineProps<{
  id: string
  data?: ProductResponse
}>()

const router = useRouter()

function getQuantity(v: ProductVendor[]) {
  let sum = 0
  for (let i of v) {
    sum += i.quantity
  }
  return sum
}

const { data, error } = useQuery({
  queryKey: ['product-images'],
  queryFn: async () =>
    api.get<ProductImageResponse[]>(`/products/images/${props.id}`),
  select: (res) => res.data,
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          // empty block
          return
        },
      })
    }
  },
  {
    immediate: true,
  },
)
</script>

<template>
  <div v-if="props.data === undefined" class="empty-product c-flex-all-center">
    <h2>There is not any product</h2>
  </div>
  <div v-else class="product-view">
    <div class="intro">
      <div v-if="data === undefined" class="img-box empty-img">
        <h2>No picture for product.</h2>
      </div>
      <div v-else class="img-box">
        <ImageFrameList :data="data" />
      </div>
      <div class="title">
        <div>
          <h1 id="title">{{ props.data.name }}</h1>
        </div>
        <div class="w-full flex flex-col gap-4 items-center">
          <p>View: {{ props.data.view_counter }}</p>
          <hr />
          <div class="price-box">
            <h2>Price: {{ props.data.price }}</h2>
            <span> | </span>
            <h2>Available quantity: {{ props.data.available_quantity }}</h2>
          </div>
          <hr />
          <h3>Published at: {{ formatDate(props.data.pub_date) }}</h3>
        </div>
      </div>
    </div>
    <div class="secondary-window">
      <div class="options c-flex-all-center">
        <div
          v-for="item of props.data.variant_detail"
          class="options-box c-flex-all-center"
        >
          <div
            v-for="[key, value] of Object.entries(item.property)"
            :key="key"
            class="options-item"
          >
            <p class="options-key">
              {{ key }}
            </p>
            <p class="options-value">
              {{ value }}
            </p>
          </div>
          <p>Price {{ item.price }}</p>
          <p>Quantity {{ getQuantity(item.vendors) }}</p>
        </div>
      </div>
      <div class="description">
        <h2 id="description">
          <span>Description</span>
        </h2>
        <p>
          {{ props.data.description }}
        </p>
      </div>
      <div class="info c-flex-all-center">
        <table id="info">
          <tr
            v-for="[key, value] of Object.entries(props.data.common_detail)"
            :key="key"
          >
            <td>
              {{ key }}
            </td>
            <td class="td-value">
              {{ value }}
            </td>
          </tr>
        </table>
      </div>
      <div class="comments">
        <!-- global components, see 'main.ts' -->
        <CommentsList
          :isParent="true"
          :relation="{
            parent: '',
            referrer: props.data.id,
          }"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.empty-product {
  @apply text-2xl font-bold p-10 size-full;
}

.product-view {
  @apply w-screen min-h-full -mt-9
    bg-linear-to-r from-c-product-view-bg-from to-c-product-view-bg-to
    text-c-product-primary-text;
}

.product-view .secondary-window {
  @apply mt-5 rounded-t-4xl pt-5 px-20
    bg-c-product-bg-secondary
    text-c-product-secondary-text;
}

.product-view .description {
  @apply w-full h-fit p-10 border-t;
}

.description h2 {
  @apply text-2xl font-bold;
}

.description p {
  @apply text-center w-full p-5;
}

.product-view .intro {
  @apply size-full p-10
    flex flex-row items-center justify-between;
}

.product-view .title {
  @apply w-full h-120 basis-5/9 max-h-180
    flex items-start justify-between flex-col gap-20
    font-bold;
}

.title h1 {
  @apply text-3xl p-10;
}

.title h2 {
  @apply text-xl;
}

.title h3 {
  @apply text-lg;
}

.title hr {
  @apply border-b-2 w-full border-c-product-hr-border;
}

.title .price-box {
  @apply w-full flex flex-row gap-2 justify-center;
}

.product-view .img-box {
  @apply size-150;
}

.product-view .empty-img {
  @apply rounded-2xl border-2
    border-common-product-view-border
    flex justify-center items-center
    text-wrap text-2xl font-bold
    overflow-hidden;
}

.product-view .options {
  @apply flex-row p-10 gap-5
    overflow-x-scroll;
}

.product-view .options-box {
  @apply flex-col border rounded-2xl
    size-40;
}

.product-view .info {
  @apply w-full h-fit p-10
    flex items-center justify-center
    border-t;
}

.product-view table {
  @apply flex flex-col gap-5 h-fit w-80/100;
}

.product-view tr {
  @apply h-fit w-full flex flex-row;
}

.product-view td {
  @apply w-full min-h-full p-5 basis-1/4 text-wrap;
}

.product-view .td-value {
  @apply border-b basis-3/4;
}
</style>
