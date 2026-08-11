<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import icons from '@/utils/icons';

const BaseIcon = defineAsyncComponent(() => import("@/components/ui/BaseIcon.vue"))
</script>

<template>
  <div class="search-box c-flex-all-center">
    <input
      class="search-input"
      id="search"
      type="search"
      placeholder="Search"
    />
    <button class="search-btn">
      <BaseIcon
        :icon="icons.ui.search.btn"
        fill-color="none"
        stroke-color="--color-search-box-icon"
      />
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.search-box {
  @apply w-full h-14 flex-row
    rounded-4xl p-1.5 bg-search-box-bg
    border-2 border-search-box-border;
}

.search-box > .search-input {
  @apply w-full h-full focus:outline-none
    mx-3 text-search-box-text text-xl;
}

.search-box > input::placeholder {
  color: var(--color-search-box-placeholder);
}

.search-box > .search-btn {
  @apply w-fit h-fit rounded-full
    bg-search-box-btn p-2 hover:cursor-pointer;
}
</style>