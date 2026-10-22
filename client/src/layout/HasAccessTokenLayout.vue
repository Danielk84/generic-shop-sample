<script setup lang="ts">
import { onBeforeMount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useStore } from '@/store'
import { setCallbackURL } from '@/utils/helper'

const store = useStore()
const router = useRouter()
const route = useRoute()

const isChecking = ref<boolean>(true)

onBeforeMount(async () => {
  if (store.hasAccessToken) {
    isChecking.value = false
    return
  }

  // The refresh cookie is shared by tabs, while the access token is kept in
  // session storage. Try to restore the session before redirecting to auth.
  if (await store.refreshAccessKey()) {
    isChecking.value = false
    return
  }

  router.replace({
    name: 'auth',
    query: {
      callback_url: setCallbackURL(route.fullPath),
    },
    replace: true,
  })
})
</script>

<template>
  <div class="has-acces">
    <p v-if="isChecking" class="has-access-loading">Loading...</p>
    <slot v-else />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.has-access-loading {
  @apply w-full min-h-screen text-xl md:text-2xl font-bold
    flex items-center justify-center;
}
</style>
