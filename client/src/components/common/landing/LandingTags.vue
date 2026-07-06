<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import { type LandingTags } from '@/types/landing';

const BaseBtn = defineAsyncComponent(() => import("@/components/common/button/BaseBtn.vue"))
const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))

const props = defineProps<{
  tags: Array<LandingTags>
}>()
</script>

<template>
  <div class="landing-tags">
    <div v-for="item of props.tags" :key="item.tag">
      <RouterLink to="/">
        <div v-if="item.isBtn"  class="base-btn-size">
          <BaseBtn>
            <span>{{ item.tag }}</span>
            <BaseIcon v-if="item.icon" :icon="item.icon" stroke-color="--color-secondary-text" fill-color="--color-primary-theme" />
          </BaseBtn>
        </div>
        <div v-else>
          {{ item.tag }}
        </div>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.landing-tags {
  @apply h-full flex items-center justify-evenly gap-32;
}
</style>
