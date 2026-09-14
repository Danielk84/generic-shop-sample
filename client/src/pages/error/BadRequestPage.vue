<script setup lang="ts">
import BadRequestSVG from '@/assets/artwork/error/400-artwork.vue'
</script>

<template>
  <div class="bad-request-page">
    <div class="artwork">
      <BadRequestSVG />
    </div>
    <h1>Why are you making a bad request?</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.bad-request-page {
  @apply w-screen min-h-screen -mt-10;
}

.bad-request-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
