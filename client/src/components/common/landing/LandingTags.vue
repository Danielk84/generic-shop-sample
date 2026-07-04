<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons-list'

const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))

const tags: Array<{
  tag: string
  isBtn?: boolean
  icon?: string
}> = [
  { tag: "Make it you'r self", isBtn: true, icon: icons.draw_1 },
  { tag: "School & Office" },
  { tag: "Home appliances" },
  { tag: "Phone case" },
  { tag: "Cards and posters" },
  { tag: "Parties" },
  { tag: "Clothing" },
] 
</script>

<template>
  <div class="landing-tags">
    <div v-for="item of tags" :key="item.tag">
        <RouterLink to="/">
        <div v-if="item.isBtn" class="btn">
          <span>{{ item.tag }}</span>
          <BaseIcon v-if="item.icon"
            :icon="item.icon"
            stroke-color="--color-btn-text-1"  />
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

.landing-tags .btn {
  @apply flex flex-row gap-1
    p-4 rounded-xl 
    text-btn-text-1 bg-primary-color
    transition delay-150 ease-out;
}
</style>