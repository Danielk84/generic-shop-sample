<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import { useStore } from '@/store'

const store = useStore()

const NavBar = defineAsyncComponent(
  () => import('@/components/common/nav-bar/NavBar.vue'),
)
const NotificationContainer = defineAsyncComponent(
  () => import('@/components/notification/NotificationContainer.vue'),
)
const LoadingPage = defineAsyncComponent(
  () => import('@/components/common/loading/LoadingPage.vue'),
)
</script>

<template>
  <div class="base">
    <header>
      <NavBar />
      <NotificationContainer />
      <div v-if="store.loading">
        <LoadingPage />
      </div>
    </header>
    <slot />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base {
  @apply font-primary-default
    bg-default-layout-bg
    text-default-layout-text
    align-middle;
}
</style>
