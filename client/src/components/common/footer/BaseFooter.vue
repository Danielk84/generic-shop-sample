<script setup lang="ts">
function getYear() {
  const d = new Date()
  return d.getFullYear()
}
</script>

<template>
  <div class="base-footer c-flex-all-center">
    Copyrigth 2025-{{ getYear() }}, All Rights Reserved, Daniel Karami.
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base-footer {
  @apply w-full h-18 font-bold
    text-sm text-footer-text
    bg-footer-bg border-t-2 border-b-2 border-footer-border;
}
</style>