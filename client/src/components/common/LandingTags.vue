<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))

const tags: Array<{
  tag: string
  isBtn?: boolean
  icon?: string
}> = [
  { tag: "Make it you'r self", isBtn: true, icon: "" },
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
        <div v-if="item.isBtn" class="btn">
          <span>{{ item.tag }}</span>
          <BaseIcon v-if="item.icon" :icon="item.icon" />
        </div>
        <div v-else>
          {{ item.tag }}
        </div>
      </div>
    </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.landing-tags {
    @apply h-full flex items-center justify-evenly;
}

.landing-tags .btn {
    @apply flex flex-row gap-1
        p-4 rounded-xl 
        text-btn-text-1 bg-primary-color
        hover:shadow-md shadow-primary-color
        transition delay-150 ease-out 
}
</style>