<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

const props = withDefaults(
  defineProps<{
    icon: string
    size?: string
    strokeColor?: string
    fillColor?: string
  }>(),
  {
    size: '24px',
    strokeColor: "--color-base-icon",
    fillColor: "--color-base-icon",
  },
)
const icon = defineAsyncComponent(() => import(`../../assets/icons/${props.icon}`))

const strokeColor = props.strokeColor == "none" ? "none" : `var(${props.strokeColor})`
const fillColor = props.fillColor == "none" ? "none" : `var(${props.fillColor})`
</script>

<template>
  <div class="base-icon">
    <icon :style="{
      width: props.size,
      height: props.size,
    }" />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base-icon svg * {
  stroke: v-bind("strokeColor");
  fill: v-bind("fillColor");
}
</style>
