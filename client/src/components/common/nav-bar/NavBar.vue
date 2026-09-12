<script setup lang="ts">
import { defineAsyncComponent, onMounted, onUnmounted, ref } from 'vue'

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
const ProfileDropdown = defineAsyncComponent(
  () => import('@/components/common/nav-bar/ProfileDropdown.vue'),
)

const navBarRef = ref<HTMLDivElement | null>(null)

function onScroll() {
  if (window.scrollY === 0) {
    navBarRef.value?.classList.remove('on-scroll')
  } else {
    navBarRef.value?.classList.add('on-scroll')
  }
}

onMounted(async () => {
  document.addEventListener('scroll', onScroll)
})
onUnmounted(async () => {
  document.removeEventListener('scroll', onScroll)
})
</script>

<template>
  <div class="nav-bar" ref="navBarRef">
    <RouterLink to="/">
      <BaseIcon :icon="icons.common.navBar.infinity" size="42px" />
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
    w-screen h-20 bg-nav-bar-bg/90
    px-10 py-5 fixed z-50
    border-b border-nav-bar-border
    backdrop-blur-2xl;
}

.nav-bar > .n-items {
  @apply flex-row gap-5 text-xl font-bold;
}

.on-scroll {
  @apply rounded-b-4xl
    transform delay-150 duration-150;
}
</style>
