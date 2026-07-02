<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const NavigationBar = defineAsyncComponent(() => import('@/components/common/NavigationBar.vue'))
const Home = defineAsyncComponent(() => import('@/components/pages/HomePage.vue'))
</script>

<template>
  <div class="app">
    <NavigationBar />
    <RouterView>
      <Home />
    </RouterView>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.app {
  @apply bg-base-bg z-10 pt-28
    w-fit min-w-screen h-fit min-h-screen;
}
</style>
