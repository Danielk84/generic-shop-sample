<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, ref, watch } from 'vue'

import api from '@/utils/api'
import { useRouter } from 'vue-router'

import icons from '@/utils/icons'
import { useStore } from '@/store'
import { PermissionType } from '@/contracts/users/request.schema'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const router = useRouter()
const store = useStore()

const showUp = ref<boolean>(false)

const onLogout = async () => {
  try {
    await api.get('auth/logout', {
      headers: { Authorization: store.getAccessToken },
    })
  } finally {
    store.logout()
    router.push({ name: 'home' })
  }
}

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
  }, 3000)
}

watch(showUp, (v) => {
  if (v) cancelClose()
})

onBeforeUnmount(() => {
  if (closeTimout) cancelClose()
})

function onClick() {
  if (store.user.isAuth) {
    showUp.value = !showUp.value
  } else {
    router.push('auth')
  }
}
</script>

<template>
  <div
    class="profile-dropdown"
    @mouseleave="scheduleClose"
    @mouseenter="cancelClose"
  >
    <button class="btn c-flex-all-center" type="button" @click="onClick()">
      <slot>Click me</slot>
      <BaseIcon v-if="store.user.isAuth" :icon="icons.common.navBar.dropdown" />
    </button>
    <div class="dropdown-content" :class="{ show: showUp }">
      <div class="content c-flex-all-center">
        <RouterLink class="item" :to="{ name: 'account' }">
          Account
        </RouterLink>
        <RouterLink class="item" :to="{ name: 'orders-list' }">
          Orders
        </RouterLink>
        <RouterLink
          v-if="store.getClaims.permission_type === PermissionType.Vendor"
          :to="{ name: 'vendor-orders' }"
          class="item"
        >
          Vendor
        </RouterLink>
        <RouterLink
          v-if="store.getClaims.permission_type === PermissionType.Admin"
          :to="{ name: 'admin' }"
          class="item"
        >
          Admin
        </RouterLink>
        <RouterLink
          v-if="store.getClaims.permission_type === PermissionType.Admin"
          :to="{ name: 'admin-issues' }"
          class="item"
        >
          Issues
        </RouterLink>
        <RouterLink v-else class="item" :to="{ name: 'issues-list' }">
          Issues
        </RouterLink>
        <button class="item logout-btn" @click="onLogout">Logout</button>
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
  @apply absolute hidden p-2
    rounded-2xl right-0 top-10
    border-2 border-(--c-v-1) dark:border-(--c-v-7)
    bg-(--c-v-7) dark:bg-(--c-v-0);
}

.profile-dropdown > .dropdown-content > .content {
  @apply flex-col gap-2;
}

.profile-dropdown .show {
  @apply block;
}

.profile-dropdown .item {
  @apply rounded-lg w-full h-10 px-9 text-center
    flex items-center justify-center
    dark:bg-(--c-v-2) bg-(--c-v-5)
    hover:brightness-110;
}

.logout-btn {
  @apply cursor-pointer font-bold text-(--c-v-11)
    hover:underline;
}
</style>
