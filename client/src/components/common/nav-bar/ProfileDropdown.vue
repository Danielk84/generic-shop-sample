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
        <RouterLink :to="{ name: 'account' }"> Account </RouterLink>
        <RouterLink :to="{ name: 'orders-list' }"> Orders </RouterLink>
        <RouterLink
          v-if="store.getClaims.permission_type === PermissionType.Vendor"
          :to="{ name: 'vendor-orders' }"
        >
          Vendor
        </RouterLink>
        <RouterLink
          v-if="store.getClaims.permission_type === PermissionType.Admin"
          :to="{ name: 'admin' }"
        >
          Admin
        </RouterLink>
        <RouterLink
          v-if="store.getClaims.permission_type === PermissionType.Admin"
          :to="{ name: 'admin-issues' }"
        >
          Issues
        </RouterLink>
        <RouterLink v-else :to="{ name: 'issues-list' }"> Issues </RouterLink>
        <button class="logout-btn" @click="onLogout">Logout</button>
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
  @apply absolute hidden px-10 py-2
    rounded-2xl right-0 top-10
    border-2 border-(--c-v-1) dark:border-(--c-v-7)
    bg-(--c-v-7) dark:bg-(--c-v-0);
}

.profile-dropdown > .dropdown-content > .content {
  @apply flex-col gap-4;
}

.profile-dropdown .show {
  @apply block;
}

.logout-btn {
  @apply cursor-pointer font-bold text-(--c-v-11)
    hover:underline;
}
</style>
