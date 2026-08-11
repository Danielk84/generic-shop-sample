<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, ref, watch } from 'vue';

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))

const showUp = ref<boolean>(false)

let closeTimout: ReturnType<typeof setTimeout> | null = null


function cancelClose() {
  if (closeTimout) {
    clearTimeout(closeTimout)
    closeTimout = null
  }
}

function scheduleClose() {
  if (closeTimout) cancelClose()
  closeTimout = setTimeout(() => {
    showUp.value = false
    closeTimout = null
  }, 3000);
}

watch(showUp, (v) => {
  if (v) cancelClose()
})

onBeforeUnmount(() => {
  if (closeTimout) cancelClose()
})
</script>

<template>
  <div
      class="profile-dropdown"
      @mouseleave="scheduleClose"  
      @mouseenter="cancelClose"
  >
    <button class="btn c-flex-all-center" @click="showUp = !showUp">
      <slot>Click me</slot>
      <BaseIcon :icon="icons.common.navBar.dropdown" />
    </button>
    <div
      class="dropdown-content"
      :class="{ 'show': showUp }">
      <div class="content c-flex-all-center">
        <RouterLink to="/">User</RouterLink>
        <RouterLink to="/">Vendor</RouterLink>
        <RouterLink to="/">Admin</RouterLink>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.profile-dropdown {
  @apply inline-block relative;
}

.profile-dropdown > .btn {
  @apply flex-row gap-1 cursor-pointer;
}

.profile-dropdown > .dropdown-content {
  @apply absolute hidden px-5 py-2
    rounded-xl right-0 top-10
    border bg-white;
}

.profile-dropdown > .dropdown-content > .content {
  @apply flex-col gap-4;
}

.profile-dropdown .show {
  @apply block;
}
</style>