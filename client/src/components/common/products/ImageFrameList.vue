<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'

import icons from '@/utils/icons'
import type { ProductImageResponse } from '@/contracts/products/response.interface'

const ImageFrameCard = defineAsyncComponent(
  () => import('@/components/card/ImageFrameCard.vue'),
)
const FullScreenImage = defineAsyncComponent(
  () => import('@/components/common/products/FullScreenImage.vue'),
)
const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const props = defineProps<{ data: ProductImageResponse[] | undefined }>()

const showUp = ref<ProductImageResponse>(
  props.data === undefined || props.data.length === 0
    ? ({} as ProductImageResponse)
    : props.data[0],
)

const fullScreen = ref<boolean>(false)
</script>

<template>
  <div v-if="props.data === undefined" class="empty-img-list c-flex-all-center">
    There are not photo for Product.
  </div>
  <div v-else class="image-frame-list c-flex-all-center">
    <div class="show-up">
      <button
        class="base-img size-full cursor-zoom-in"
        @click="fullScreen = true"
      >
        <ImageFrameCard :img="showUp.img_path" :alt="showUp.id" />
        <div class="full-screen-btn">
          <BaseIcon
            :icon="icons.common.products.fullScreen"
            fill-color="--c-v-11"
            stroke-color="--c-v-11"
            fill-dark-color="--c-v-11"
            stroke-dark-color="--c-v-11"
            size="32px"
          />
        </div>
      </button>
    </div>
    <div class="list">
      <div v-for="i in props.data" :key="i.img_path" class="base-img item">
        <button class="btn" @click="showUp = i">
          <ImageFrameCard :img="i.img_path" :alt="i.id" />
        </button>
      </div>
    </div>
    <div v-if="fullScreen">
      <FullScreenImage
        :img="showUp.img_path"
        :alt="showUp.id"
        @destroy="fullScreen = false"
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.empty-img-list {
  @apply size-150 rounded-4xl border-4
    border-t-(--c-v-12)
    border-r-(--c-v-13)
    border-b-(--c-v-14)
    border-l-(--c-v-15)
    bg-(--c-v-1)
    text-(--c-v-11) font-bold text-2xl;
}

.image-frame-list {
  @apply w-150 h-150 flex-col;
}

.image-frame-list > .list {
  @apply w-full h-20/100
    overflow-y-hidden overflow-x-scroll
    flex flex-row flex-nowrap items-center gap-4;
}

.image-frame-list > .list > .item {
  @apply size-25 shrink-0;
}

.image-frame-list .btn {
  @apply size-full cursor-pointer;
}

.image-frame-list > .show-up {
  @apply w-full h-80/100;
}

.image-frame-list .base-img {
  @apply border rounded-xl border-(--c-v-10)
    hover:brightness-80 relative;
}

.image-frame-list .full-screen-btn {
  @apply size-full flex justify-end items-end
    absolute bottom-0 z-10 p-4;
}
</style>
