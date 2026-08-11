<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons'
import type { ImageFrameCardProps } from '@/components/card/types'

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))
const ImageFrameCard = defineAsyncComponent(() => import('@/components/card/ImageFrameCard.vue'))

const props = defineProps<ImageFrameCardProps>()
const emits = defineEmits<{
  (e: 'destroy'): void
}>()
</script>

<template>
  <div class="full-screen">
    <div class="tools-bar">
      <button
        class="close-btn"
        @click="emits('destroy')"  
      >
        <BaseIcon :icon="icons.pages.products.close" />
      </button>
    </div>
    <div class="item">
      <ImageFrameCard :img="props.img" :alt="props.alt" />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.full-screen {
  @apply w-screen h-screen z-80 fixed
    backdrop-blur-lg backdrop-brightness-75 overflow-hidden inset-0;
}

.full-screen > .tools-bar {
  @apply w-full h-9 border-b
    border-img-frame-list-border px-10
    flex flex-row items-center justify-end;
}

.full-screen > .tools-bar > .close-btn {
  @apply cursor-pointer;
}

.full-screen > .item {
  @apply size-full p-20 object-cover;
}
</style>