<script setup lang="ts">
import NotFoundSVG from '@/assets/artwork/error/404-artwork.vue'
</script>

<template>
  <div class="not-found-page">
    <div class="artwork">
      <NotFoundSVG />
    </div>
    <h1>Page or content not found.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.not-found-page {
  @apply w-full min-h-screen p-4 md:p-10 md:-mt-50;
}

.not-found-page h1 {
  @apply text-4xl md:text-7xl font-extrabold w-full p-6 md:p-10 text-center
    text-(--c-v-11);
}
</style>
