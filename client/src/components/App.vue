<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const NavigationBar = defineAsyncComponent(() => import('@/components/common/NavigationBar.vue'))
</script>

<template>
  <div class="app">
    <NavigationBar />
    <RouterView />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.app {
  @apply bg-default-bg font-default-font
    z-10 pt-28 max-w-screen min-h-screen;
}
</style>
