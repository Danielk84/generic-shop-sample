<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import type { ImageFrame } from '@/types/cards';
import type { DestroyEmits } from '@/types/btn';
import icons from '@/utils/icons-list';

const CloseBtn = defineAsyncComponent(() => import("@/components/common/button/CloseBtn.vue"))

const props = defineProps<ImageFrame>()
const emits = defineEmits<DestroyEmits>()
</script>

<template>
  <div class="full-screen-image-frame">
    <div class="nav-bar">
      <CloseBtn @destroy="emits('destroy')" :icon="icons.exitfullScreen" />
    </div>
    <div class="image-frame c-flex-all-center">
      <img :src="props.src" :alt="props.alt">
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.full-screen-image-frame {
  @apply absolute inset-0 z-50 w-full h-full overflow-hidden backdrop-blur-lg;
}

.full-screen-image-frame .nav-bar {
  @apply w-full h-fit flex items-center justify-end
    border-b border-baseline p-2;
}

.full-screen-image-frame .image-frame {
  @apply w-full h-full;
}

.full-screen-image-frame .image-frame img {
  @apply max-w-full max-h-full object-contain p-2;
}
</style>