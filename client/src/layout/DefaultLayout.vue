<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'

const store = useStore()
const notification = useNotificationStore()

const NavBar = defineAsyncComponent(
  () => import('@/components/common/nav-bar/NavBar.vue'),
)
const NotificationContainer = defineAsyncComponent(
  () => import('@/components/notification/NotificationContainer.vue'),
)
const LoadingPage = defineAsyncComponent(
  () => import('@/components/common/loading/LoadingPage.vue'),
)

notification.initTimout()
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
    bg-(--c-v-7) dark:bg-(--c-v-0)
    text-(--c-v-1) dark:text-(--c-v-7)
    align-middle;
}
</style>
