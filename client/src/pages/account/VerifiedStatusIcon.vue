<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons'

const props = withDefaults(defineProps<{ status?: boolean }>(), {
  status: false,
})

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
</script>

<template>
  <BaseIcon
    v-if="props.status"
    :icon="icons.pages.account.verified"
    size="32px"
    fill-color="none"
    fill-dark-color="none"
    stroke-color="--c-v-14"
    stroke-dark-color="--c-v-14"
  />
  <BaseIcon
    v-else
    :icon="icons.pages.account.notVerified"
    size="32px"
    fill-color="--c-v-11"
    fill-dark-color="--c-v-11"
    stroke-color="--c-v-11"
    stroke-dark-color="--c-v-11"
  />
</template>
