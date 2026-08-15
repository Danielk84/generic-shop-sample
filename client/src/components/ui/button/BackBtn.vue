<script setup lang="ts">
import { defineAsyncComponent } from 'vue';
import { useRouter, type RouteLocationRaw } from 'vue-router';

import icons from '@/utils/icons'
import type { SVGIcon } from '@/components/ui/types';

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))

const props = defineProps<{ to: RouteLocationRaw, icon?: SVGIcon }>()
const router = useRouter()
</script>

<template>
  <div>
    <button
      class="back-btn"
      @click="router.push(props.to)"
    >
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