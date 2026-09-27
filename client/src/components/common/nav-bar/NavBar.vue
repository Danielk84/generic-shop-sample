<script setup lang="ts">
import { defineAsyncComponent, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
const ProfileDropdown = defineAsyncComponent(
  () => import('@/components/common/nav-bar/ProfileDropdown.vue'),
)

const router = useRouter()
const store = useStore()
const navBarRef = ref<HTMLDivElement | null>(null)

async function openBasket() {
  if (!store.hasAccessToken) {
    router.push({ name: 'auth' })
    return
  }
  try {
    const res = await api.post<{ order_id: string }>(
      'orders/',
      {},
      { headers: { Authorization: store.getAccessToken } },
    )
    router.push({ name: 'basket', params: { id: res.data.order_id } })
  } catch (err) {
    errorStatusHandler(err as Error, router)
  }
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
    store.setDarkTheme(true)
    document.documentElement.classList.add('dark')
  } else {
    store.setDarkTheme(false)
    document.documentElement.classList.remove('dark')
  }
}

function toggleTheme() {
  store.setDarkTheme(!store.isDarkTheme)
  if (store.isDarkTheme) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  }
}

function onScroll() {
  if (window.scrollY === 0) {
    navBarRef.value?.classList.remove('on-scroll')
  } else {
    navBarRef.value?.classList.add('on-scroll')
  }
}

onMounted(async () => {
  initTheme()
  document.addEventListener('scroll', onScroll)

  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
      if (e.matches) {
        document.documentElement.classList.add('dark')
        store.setDarkTheme(true)
      } else {
        document.documentElement.classList.remove('dark')
        store.setDarkTheme(false)
      }
    }
  })
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
      <RouterLink :to="{ name: 'categories' }">
        <span>Categories</span>
      </RouterLink>
      <RouterLink to="/">
        <span>Contact Us</span>
      </RouterLink>
      <RouterLink :to="{ name: 'products-list' }">
        <span>Products</span>
      </RouterLink>
    </div>
    <div class="n-items c-flex-all-center">
      <button class="base-btn" @click="toggleTheme" aria-label="Toggle theme">
        <BaseIcon
          v-if="store.isDarkTheme"
          :icon="icons.ui.theme.sun"
          size="24px"
        />
        <BaseIcon v-else :icon="icons.ui.theme.moon" size="24px" />
      </button>
      <button class="base-btn" @click="openBasket">
        <BaseIcon :icon="icons.common.navBar.shoppingBag" />
      </button>
      <ProfileDropdown>
        <BaseIcon :icon="icons.common.navBar.profile" />
      </ProfileDropdown>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.nav-bar {
  @apply flex flex-row justify-between items-center
    w-screen h-20
    bg-(--c-v-6)/90 dark:bg-(--c-v-3)/80
    px-10 py-5 fixed z-50
    border-b border-(--c-v-1) dark:border-(--c-v-7)
    backdrop-blur-2xl;
}

.nav-bar > .n-items {
  @apply flex-row gap-5 text-xl font-bold;
}

.base-btn {
  @apply p-2 rounded-lg
    bg-(--c-v-6) dark:bg-(--c-v-1)
    border border-(--c-v-1) dark:border-(--c-v-7)
    hover:brightness-110 transition-colors cursor-pointer;
}

.on-scroll {
  @apply rounded-b-4xl
    transform delay-150 duration-150;
}

@media (max-width: 768px) {
  .nav-bar {
    @apply px-4 h-16;
  }

  .nav-bar > .n-items:first-of-type {
    @apply hidden;
  }

  .nav-bar > .n-items {
    @apply gap-2 text-base;
  }
}
</style>
