<script setup lang="ts">
import { type StyleValue, computed } from 'vue';

import type { Btn } from "@/types/btn";
import { getCssVar } from '@/utils/css-helper';

const props = withDefaults(
  defineProps<Btn>(),
  {
    style: () => ({
      backgroundColor: "--color-primary-theme",
      outlineColor: "none",
      outlineWidth: "0px",
      color: "--color-secondary-text"
    }),
  }
)
const emits = defineEmits<{
  (e: 'click', event: MouseEvent): void,
}>()

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
    @click="emits('click', $event)"
    :style="style">
      <slot>empty!</slot>
  </button>
</template>

<style>
@reference "@/styles/index.css";

.base-btn {
  @apply rounded-lg text-nowrap font-bold hover:brightness-125
    transition delay-150 ease-out
    w-full h-full cursor-pointer outline-solid;
}
</style>
