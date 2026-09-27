<script setup lang="ts">
import { defineAsyncComponent, computed } from 'vue'

import { getCssVar } from '@/utils/helper'
import { useStore } from '@/store'
import type { Icon } from '@/components/ui/types'

const props = withDefaults(defineProps<Icon>(), {
  size: '24px',

  strokeColor: '--c-v-2',
  strokeDarkColor: '--c-v-7',

  fillColor: '--c-v-2',
  fillDarkColor: '--c-v-7',
})

const store = useStore()

const icon = defineAsyncComponent(
  () => import(/* @vite-ignore */ `../../assets/icons/${props.icon}`),
)

const style = computed(() => ({
  strokeColor: getCssVar(
    store.isDarkTheme ? props.strokeDarkColor : props.strokeColor,
  ),
  fillColor: getCssVar(
    store.isDarkTheme ? props.fillDarkColor : props.fillColor,
  ),
}))
</script>

<template>
  <div class="base-icon">
    <icon
      :style="{
        width: props.size,
        height: props.size,
      }"
    />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base-icon > svg > * {
  stroke: v-bind('style.strokeColor');
  fill: v-bind('style.fillColor');
}
</style>
