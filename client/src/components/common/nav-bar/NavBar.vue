<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))
const ProfileDropdown = defineAsyncComponent(() => import('@/components/common/nav-bar/ProfileDropdown.vue'))
</script>

<template>
  <div class="nav-bar">
    <RouterLink to="/">
      <BaseIcon :icon="icons.common.navBar.infinity" size="42px"/>
    </RouterLink>
    <div class="n-items c-flex-all-center">
      <RouterLink to="/">
        <span>Home</span>
      </RouterLink>
      <div>
        <span>Categories</span>
      </div>
      <RouterLink to="/">
        <span>Contact Us</span>
      </RouterLink>
      <RouterLink to="/">
        <span>Products</span>
      </RouterLink>
    </div>
    <div class="n-items c-flex-all-center">
      <ProfileDropdown>
        <BaseIcon :icon="icons.common.navBar.profile" />
      </ProfileDropdown>
      <RouterLink to="/">
        <BaseIcon :icon="icons.common.navBar.shoppingBag" />
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.nav-bar {
  @apply flex flex-row justify-between items-center
    w-screen h-20 bg-nav-bar-bg
    px-10 py-5 fixed z-50
    border-b border-nav-bar-border;
}

.nav-bar > .n-items {
  @apply flex-row gap-5 text-xl font-bold;
}
</style>