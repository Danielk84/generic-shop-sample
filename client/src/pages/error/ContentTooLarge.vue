<script setup lang="ts">
import ContentTooLargeSVG from '@/assets/artwork/error/413-artwork.vue'
</script>

<template>
  <div class="bad-request-page">
    <div class="artwork">
      <ContentTooLargeSVG />
    </div>
    <h1>Why data is so big?</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.bad-request-page {
  @apply w-full min-h-screen md:mt-10;
}

.bad-request-page .artwork {
  @apply max-md:max-h-110;
}

.bad-request-page .artwork {
  @apply w-full flex justify-center;
}

.bad-request-page h1 {
  @apply text-4xl md:text-7xl font-extrabold w-full p-6 md:p-10 text-center
    text-(--c-v-11);
}
</style>
