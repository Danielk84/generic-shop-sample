<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const NavigationBar = defineAsyncComponent(() => import('@/components/common/NavigationBar.vue'))
const Home = defineAsyncComponent(() => import('@/components/pages/HomePage.vue'))
</script>

<template>
  <div>
    <NavigationBar />
    <RouterView>
      <Home />
    </RouterView>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
