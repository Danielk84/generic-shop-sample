<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

import { imageOnLoadHook } from '@/components/card/hooks'
import type { ImageFrameCardProps } from '@/components/card/types'

const props = defineProps<ImageFrameCardProps>()

const imgRef = ref<HTMLImageElement | null>(null)
const bgRef = ref<HTMLDivElement | null>(null)

const onload = imageOnLoadHook(imgRef, bgRef)

onMounted(async () => {
  imgRef.value?.addEventListener('load', onload)
})
onUnmounted(async () => {
  imgRef.value?.removeEventListener('load', onload)
})
</script>

<template>
  <div
    class="img-frame"
    ref="bgRef"
  >
    <div class="img-box c-flex-all-center">
      <img
        v-if="typeof img === 'string'"
        ref="imgRef"
        :src="props.img"
        :alt="props.alt"
        loading="lazy"
      />
    </div>
    <div v-if="typeof tag == 'string'" class="tag">
      <span>{{ props.tag }}</span>
    </div>
  </div>
</template>

<style scoped>
@reference '@/styles/index.css';

.img-frame {
  @apply flex-col w-full h-full z-0
    bg-img-frame-card-bg rounded-xl
    relative isolate overflow-hidden;
}

.img-frame > .img-box {
  @apply z-10 absolute inset-0 backdrop-blur-2xl;
}

.img-frame > .img-box > img {
  @apply w-full h-full object-contain;
}

.img-frame .tag {
  @apply w-fit h-fit absolute inset-0 p-2 m-4 z-20
    bg-img-frame-card-tag-bg rounded-xl brightness-125
    text-img-frame-card-tag-text;
}
</style>
