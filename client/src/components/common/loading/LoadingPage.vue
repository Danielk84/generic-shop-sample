<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
</script>

<template>
  <div class="loading-page c-flex-all-center">
    <div class="loading-icon">
      <BaseIcon
        :icon="icons.common.loading.loading"
        size="80px"
        stroke-color="--color-loading-color"
        fill-color="none"
      />
    </div>
    <h2>Try Loading Content...</h2>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.loading-page {
  @apply w-screen h-screen backdrop-blur-xl
    absolute inset-0 z-70
    text-xl font-bold
    gap-5 flex-col
    text-loading-color;
}

.loading-page .loading-icon {
  @apply motion-safe:animate-spin;
}
</style>
