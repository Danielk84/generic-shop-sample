<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from "@/utils/icons-list"

const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
</script>

<template>
  <div class="base">
    <button>
      <BaseIcon :icon="icons.search"
        fill-color="none"
        stroke-color="--color-primary-color"/>
    </button>
    <input type="text" id="search" placeholder="search" class="search-box" />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base {
  @apply flex flex-row items-center justify-center gap-2
        min-w-xl h-fit p-2 rounded-xl
        border border-navigation-bar;
}

.base input::placeholder {
  color: var(--color-primary-color);
}

.search-box {
  @apply w-full p-1 focus:outline-0;
}
</style>
