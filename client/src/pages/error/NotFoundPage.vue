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
  @apply w-screen min-h-screen -mt-50;
}

.not-found-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
