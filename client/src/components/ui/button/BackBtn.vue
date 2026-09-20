<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'

import icons from '@/utils/icons'
import type { SVGIcon } from '@/components/ui/types'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const props = defineProps<{ icon?: SVGIcon; pageName?: string }>()
const router = useRouter()

function onClick(event: MouseEvent) {
  event.preventDefault()
  if (props.pageName === undefined) {
    router.back()
  }
  if (props.pageName === 'back-2') {
    router.go(-2)
  } else {
    router.push({ name: props.pageName })
  }
}
</script>

<template>
  <div>
    <button class="back-btn" @click="onClick($event)">
      <BaseIcon
        :icon="icons.ui.button.back"
        :size="props.icon?.size"
        :stroke-color="props.icon?.strokeColor"
        :fill-color="props.icon?.fillColor"
      />
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.back-btn {
  @apply size-fit cursor-pointer;
}
</style>
