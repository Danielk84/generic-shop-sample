<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import icons from '@/utils/icons-list';

const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))

const props = defineProps<{
  msg: string
  seeAlsoLink?: string
}>()
</script>

<template>
  <div class="section-title">
    <h2 class="title">
      <BaseIcon :icon="icons.medalRibbonsStar" size="50px" />
      <span>
        {{ props.msg }}
      </span>
    </h2>
    <div class="gap-line"></div>
    <div v-if="props.seeAlsoLink">
      <RouterLink :to="props.seeAlsoLink">
        <span class="see-more">
          See more
        </span>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.section-title {
  @apply flex flex-row gap-6 px-10 items-center justify-stretch
    w-full;
}

.section-title .title {
  @apply flex flex-row gap-1 text-nowrap
    font-bold text-4xl;
}

.section-title .gap-line {
  @apply h-5/10 w-full border-b border-navigation-bar;
}

.section-title .see-more {
  @apply text-primary-color text-nowrap;
}

</style>