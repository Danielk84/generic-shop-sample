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
    text-sm text-(--c-v-7)
    bg-(--c-v-1) border-t-2 border-b-2 border-(--c-v-7);
}
</style>
