<script setup lang="ts">
import UnprocessableContentSVG from '@/assets/artwork/error/422-artwork.vue'
</script>

<template>
  <div class="unprocessable-content-page c-flex-all-center">
    <div class="artwork">
      <UnprocessableContentSVG />
    </div>
    <h1>We couldn't process the provided content.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.unprocessable-content-page {
  @apply w-full min-h-screen flex-col max-sm:-mt-20 md:mt-10 overflow-hidden;
}

.unprocessable-content-page .artwork {
  @apply max-w-full;
}

.unprocessable-content-page h1 {
  @apply text-4xl md:text-7xl font-extrabold w-full p-6 md:p-10 text-center
    text-(--c-v-11);
}
</style>
