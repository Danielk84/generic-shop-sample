<script setup lang="ts">
import InternalServerErrorSVG from '@/assets/artwork/error/500-artwork.vue'
</script>

<template>
  <div class="internal-server-error-page c-flex-all-center">
    <div class="artwork">
      <InternalServerErrorSVG />
    </div>
    <h1>Sorry, Internal server error.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.internal-server-error-page {
  @apply w-screen min-h-screen flex-col overflow-hidden;
}

.internal-server-error-page .artwork {
  @apply max-w-full;
}

.internal-server-error-page h1 {
  @apply text-4xl md:text-7xl font-extrabold w-full p-6 md:p-10 text-center
    text-(--c-v-11);
}
</style>
