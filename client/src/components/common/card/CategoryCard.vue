<script setup lang="ts">
import { type StyleValue, computed } from "vue";

import { getCssUrl } from "@/utils/css-helper";

const props = defineProps<{
  backgroundImage?: string
}>()

const style = computed<StyleValue>(() => ({
  backgroundImage: getCssUrl(props.backgroundImage),
}))
</script>

<template>
  <div class="category-card c-flex-all-center default-shadow-hover">
    <div class="banner default-banner-bg" :style="style"></div>
    <div class="tag">
      <RouterLink to="/">
        <span>
          <slot>NO Content!</slot>
        </span>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.category-card {
  @apply flex-col gap-4 p-2 rounded-2xl
    border border-baseline
    w-60 h-42;
}

.category-card .banner {
  @apply rounded-lg
  w-98/100 h-10/12;
}

.category-card .tag {
  @apply font-bold;
}
</style>
