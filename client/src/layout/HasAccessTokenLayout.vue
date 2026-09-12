<script setup lang="ts">
import { onBeforeMount } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useStore } from '@/store'
import { setCallbackURL } from '@/utils/helper'

const store = useStore()
const router = useRouter()
const route = useRoute()

onBeforeMount(() => {
  if (!store.hasAccessToken) {
    router.replace({
      name: 'auth',
      query: {
        callback_url: setCallbackURL(route.fullPath),
      },
      replace: true,
    })
  }
})
</script>

<template>
  <div class="has-acces">
    <slot />
  </div>
</template>
