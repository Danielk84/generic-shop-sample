<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons-list'

const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
const SearchBox = defineAsyncComponent(() => import('@/components/common/SearchBox.vue'))
</script>

<template>
  <nav class="top-nav">
    <nav class="sub-nav c-flex-all-center">
      <RouterLink to="/" class="right-separator px-5">
        <BaseIcon :icon="icons.shopping" />
      </RouterLink>
      <RouterLink v-if="$route.name !== 'auth'" :to="{ name: 'auth' }" class="item">
        <span>Login | Sing Up</span>
        <BaseIcon :icon="icons.login" />
      </RouterLink>
    </nav>
    <nav class="sub-nav c-flex-all-center">
      <div>
        <SearchBox />
      </div>
      <RouterLink to="/" class=""> logo </RouterLink>
    </nav>
  </nav>
</template>

<style scoped>
@reference "@/styles/index.css";

.top-nav {
  @apply flex flex-row items-center justify-between
    z-50 fixed top-0 backdrop-blur-md
    border-b border-baseline
    w-full h-20 p-2 bg-default-bg/92;
}

.top-nav .sub-nav {
  @apply gap-5;
}

.top-nav .item {
  @apply flex flex-row items-center justify-evenly
    border bg-default-bg
    p-3 gap-2 rounded-xl;
}
</style>
