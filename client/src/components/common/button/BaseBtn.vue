<script setup lang="ts">
import { type StyleValue, computed } from 'vue';

import { type BaseBtn } from "@/types/btn";
import { getCssVar } from '@/utils/css-utils';

const props = withDefaults(
  defineProps<BaseBtn>(),
  {
    click: () => {},
    style: () => ({
      backgroundColor: "--color-primary-theme",
      outlineColor: "none",
      outlineWidth: "0px",
      color: "--color-secondary-text"
    }),
  }
)

const style = computed<StyleValue>(() => ({
  backgroundColor: getCssVar(props.style.backgroundColor as string),
  outlineColor: getCssVar(props.style.outlineColor as string),
  outlineWidth: props.style.outlineWidth as string,
  color: getCssVar(props.style.color as string),
}))
</script>

<template>
  <button
    class="base-btn c-flex-all-center"
    @click="props.click($event)"
    :style="style">
      <slot>empty!</slot>
  </button>
</template>

<style>
@reference "@/styles/index.css";

.base-btn {
  @apply rounded-2xl p-2 text-nowrap font-bold;
}
</style>
