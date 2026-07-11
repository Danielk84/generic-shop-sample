<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';

import type { ImageFrame } from '@/types/cards';
import { imageOnLoadHook } from '@/hooks/cards';

const props = defineProps<ImageFrame>()
const emits = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const imgRef = ref<HTMLImageElement | null>(null)
const bgRef = ref<HTMLDivElement | null>(null)

const onload = imageOnLoadHook(imgRef, bgRef)

onMounted(async () => {
  imgRef.value?.addEventListener("load", onload)
})
onUnmounted(async () => {
  imgRef.value?.removeEventListener("load", onload)
})
</script>

<template>
  <div  class="image-frame default-banner-bg"
    @click="emits('click', $event)">
    <div ref="bgRef" class="sub-image-frame"></div>
    <div class="base-img c-flex-all-center">
      <img ref="imgRef" :src="props.src" :alt="props.alt" loading="lazy"/>
    </div>
    <div class="image-frame-box">
      <slot />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.image-frame {
  @apply rounded-lg w-full h-full relative isolate overflow-hidden;
}

.image-frame .sub-image-frame {
  @apply w-full h-full z-10 absolute inset-0;
}

.image-frame .base-img {
  @apply w-full h-full z-20 overflow-hidden absolute inset-0 backdrop-blur-lg;
}

.image-frame .base-img img {
  @apply max-w-full max-h-full object-contain;
}

.image-frame .image-frame-box {
  @apply w-full h-full z-30 absolute inset-0;
}
</style>