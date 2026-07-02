<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from "@/utils/icons-list"

const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
</script>

<template>
  <div class="search-box">
    <button>
      <BaseIcon :icon="icons.search"
        fill-color="none"
        stroke-color="--color-primary-color"/>
    </button>
    <input type="text" id="search" placeholder="search" class="search-box-input" />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.search-box {
  @apply flex flex-row items-center justify-center gap-2
    min-w-xl h-fit p-2 rounded-xl relative
    border border-navigation-bar;
}

.search-box input::placeholder {
  color: var(--color-primary-color);
}

.search-box-input {
  @apply w-full p-1 focus:outline-0;
}
</style>
