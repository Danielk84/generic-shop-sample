<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))
</script>

<template>
  <div class="subscribe-box">
    <input
      id="subscribe-input"
      type="text"
      placeholder="person@email.com"
    />
    <button class="subscribe-btn c-flex-all-center">
      <BaseIcon
        :icon="icons.common.footer.send"
        stroke-color="--color-footer-icon"
        fill-color="--color-footer-icon"
      />
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.subscribe-box {
  @apply w-full h-14 flex flex-row
    rounded-4xl p-1.5 bg-footer-input
    border-2 border-footer-border;
}

.subscribe-box > input {
  @apply w-full h-full focus:outline-none
    mx-3 text-footer-text text-xl;
}

.subscribe-box > input::placeholder {
  color: var(--color-footer-placeholder);
}

.subscribe-box > .subscribe-btn {
  @apply rounded-full w-fit h-fit
    bg-footer-btn p-2 hover:cursor-pointer;
}
</style>