<script setup lang="ts">
import { defineAsyncComponent } from "vue";

const LandingTags = defineAsyncComponent(() => import("@/components/common/landing/LandingTags.vue"))
const LandingBanner = defineAsyncComponent(() => import("@/components/common/landing/LandingBanner.vue"))
</script>

<template>
  <div class="base-home">
    <LandingTags />
    <LandingBanner />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base-home {
  @apply flex flex-col items-center gap-7
    w-full min-h-140;
}
</style>
