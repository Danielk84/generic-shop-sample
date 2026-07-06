<script setup lang="ts">
import { computed, type StyleValue } from "vue";

import { getCssUrl } from "@/utils/css-helper";

const props = defineProps<{
  to: string
  backgroundImage?: string
}>()

const style = computed<StyleValue>(() => ({
  backgroundImage: getCssUrl(props.backgroundImage),
}))
</script>

<template>
  <RouterLink :to="props.to">
    <div class="product-v-card default-shadow-hover">
      <div class="banner default-banner-bg" :style="style"></div>
      <slot>No Content!</slot>
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.product-v-card {
  @apply flex flex-col items-center gap-6 p-3 rounded-2xl
    border border-baseline
    w-95 h-145;
}

.product-v-card .banner {
  @apply w-98/100 h-9/12 rounded-lg;
}
</style>
