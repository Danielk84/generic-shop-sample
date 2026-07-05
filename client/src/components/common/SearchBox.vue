<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons-list'

const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
</script>

<template>
  <button class="c-flex-all-center search-box" @click="() => {}">
    <BaseIcon :icon="icons.search" fill-color="none" stroke-color="--color-primary-theme" />
    <input type="search" id="search" placeholder="search" class="search-box-input" />
  </button>
</template>

<style scoped>
@reference "@/styles/index.css";

.search-box {
  @apply flex-row gap-2
    min-w-xl h-fit p-2 rounded-xl relative
    border border-baseline bg-default-bg;
}

.search-box input::placeholder {
  color: var(--color-primary-theme);
}

.search-box-input {
  @apply w-full p-1 focus:outline-0;
}
</style>
