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
  @apply w-screen min-h-screen mt-10;
}

.bad-request-page .artwork {
  @apply w-full flex justify-center;
}

.bad-request-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
