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
  @apply w-screen min-h-screen flex-col;
}

.internal-server-error-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
