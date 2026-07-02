<script setup lang="ts">
import { defineAsyncComponent } from "vue";

const LandingTags = defineAsyncComponent(() => import("@/components/common/LandingTags.vue"))
</script>

<template>
  <div class="base-home">
    <LandingTags />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base-home {
  @apply h-full;
}
</style>
