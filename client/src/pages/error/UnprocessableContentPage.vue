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
  @apply w-screen min-h-screen flex-col mt-10;
}

.unprocessable-content-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
