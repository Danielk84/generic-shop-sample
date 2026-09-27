<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
</script>

<template>
  <div class="subscribe-box">
    <input id="subscribe-input" type="text" placeholder="person@email.com" />
    <button class="subscribe-btn c-flex-all-center">
      <BaseIcon
        :icon="icons.common.footer.send"
        stroke-color="--c-v-7"
        fill-color="--c-v-7"
      />
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.subscribe-box {
  @apply w-full h-14 flex flex-row
    rounded-4xl p-1.5 bg-(--c-v-3)
    border-2 border-(--c-v-7);
}

.subscribe-box > input {
  @apply w-full h-full focus:outline-none
    mx-3 text-(--c-v-7) text-xl;
}

.subscribe-box > input::placeholder {
  color: var(--c-v-5);
}

.subscribe-box > .subscribe-btn {
  @apply rounded-full w-fit h-fit
    bg-(--c-v-1) p-2 hover:cursor-pointer;
}
</style>
