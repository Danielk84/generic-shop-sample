<script setup lang="ts">
import { defineAsyncComponent, computed } from 'vue'

import { getCssVar } from '@/utils/helper';
import type { Icon } from '@/components/ui/types';

const props = withDefaults(
  defineProps<Icon>(),
  {
    size: '24px',
    strokeColor: '--color-default-icon',
    fillColor: '--color-default-icon',
  },
)

const icon = defineAsyncComponent(() => import( /* @vite-ignore */ `../../assets/icons/${props.icon}`))

const style = computed(() => ({
  strokeColor: getCssVar(props.strokeColor),
  fillColor: getCssVar(props.fillColor),
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
