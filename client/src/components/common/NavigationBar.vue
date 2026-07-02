<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from "@/utils/icons-list"

const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
const SearchBox = defineAsyncComponent(() => import('@/components/common/SearchBox.vue'))
</script>

<template>
  <nav class="base">
    <RouterLink to="/" class="item">
      <span>Shopping card</span>
      <BaseIcon :icon="icons.shopping" />
    </RouterLink>
    <RouterLink to="/" class="item">
      <span>Login | Sing Up</span>
      <BaseIcon :icon="icons.login" />
    </RouterLink>
    <div>
      <SearchBox />
    </div>
    <RouterLink to="/" class=""> logo </RouterLink>
  </nav>
</template>

<style scoped>
@reference "@/styles/index.css";

.base {
  @apply flex flex-row items-center justify-evenly
        border-b border-navigation-bar
        w-full h-24 p-2;
}

.item {
  @apply flex flex-row items-center justify-evenly
        shadow-shadown-hover
        hover:shadow-xs p-3 gap-2 rounded-xl
        transition delay-150 ease-in;
}
</style>
