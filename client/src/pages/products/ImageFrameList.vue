<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'

import icons from '@/utils/icons'
import type { ImageFrameListProps } from '@/pages/products/types'
import type { ImageFrameCardProps } from '@/components/card/types'

const ImageFrameCard = defineAsyncComponent(() => import('@/components/card/ImageFrameCard.vue'))
const FullScreenImage = defineAsyncComponent(() => import('@/pages/products/FullScreenImage.vue'))
const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))

const props = withDefaults(
  defineProps<{ data: ImageFrameListProps }>(),
  {
    data: (): ImageFrameListProps => ({
      images: [],
    }),
  }
)

const showUp = ref<ImageFrameCardProps>(
  props.data.images.length === 0 ?
    {} as ImageFrameCardProps:
    props.data.images[0]
)

const fullScreen = ref<boolean>(false)
</script>

<template>
  <div class="image-frame-list c-flex-all-center">
    <div class="list">
      <div v-for="i in props.data.images" :key="i.img" class="base-img item">
        <button
          class="btn"
          @click="showUp = i"
        >
          <ImageFrameCard :img="i.img" :alt="i.alt" />
        </button>
      </div>
    </div>
    <div class="show-up">
      <button
        class="base-img btn"
        @click="fullScreen = true"
      >
        <ImageFrameCard :img="showUp.img" :alt="showUp.alt" />
        <div class="full-screen-btn">
          <BaseIcon
            :icon="icons.pages.products.fullScreen"
            fill-color="--color-img-frame-list-icon"
            stroke-color="--color-img-frame-list-icon"
            size="32px"
          />
        </div>
      </button>
    </div>
    <div v-if="fullScreen">
      <FullScreenImage
        :img="showUp.img"
        :alt="showUp.alt"
        @destroy="fullScreen = false"
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.image-frame-list {
  @apply w-150 h-150 flex-row;
}

.image-frame-list > .list {
  @apply w-20/100 h-full
    overflow-x-hidden overflow-y-scroll
    flex flex-col flex-nowrap items-center gap-4;
}

.image-frame-list > .list > .item {
  @apply size-25 shrink-0;
}

.image-frame-list .btn {
  @apply size-full cursor-pointer;
}

.image-frame-list > .show-up {
  @apply w-80/100 h-full;
}

.image-frame-list .base-img {
  @apply border rounded-xl border-img-frame-list-border
    hover:brightness-80 relative;
}

.image-frame-list .full-screen-btn {
  @apply size-full flex justify-end items-end
    absolute bottom-0 z-10 p-4;
}
</style>