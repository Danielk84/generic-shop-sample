<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const NavBar = defineAsyncComponent(
  () => import('@/components/common/nav-bar/NavBar.vue'),
)
const NotificationContainer = defineAsyncComponent(
  () => import('@/components/notification/NotificationContainer.vue'),
)
</script>

<template>
  <div class="base">
    <header>
      <NavBar />
      <NotificationContainer />
    </header>
    <slot />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base {
  @apply font-primary-default;
}
</style>
