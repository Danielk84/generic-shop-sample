<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { defineAsyncComponent, watch } from 'vue'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
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
      errorStatusHandler(err, router)
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
      <div></div>
    </div>
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
    <div class="info">
      <div
        v-for="[key, value] of Object.entries(props.data.common_detail)"
        :key="key"
        class="info-item"
      >
        <p class="info-item">
          {{ key }}
        </p>
        <p class="info-value">
          {{ value }}
        </p>
      </div>
    </div>
    <div class="comments"></div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.empty-product {
  @appy text-2xl font-bold p-10 size-full;
}

.product-view {
  @apply w-screen min-h-full;
}

.product-view .intro {
  @apply flex flex-row items-center justify-between p-8;
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
  @apply flex flex-col;
}
</style>
