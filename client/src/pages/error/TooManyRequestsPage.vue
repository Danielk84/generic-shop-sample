<script setup lang="ts">
import TooManyRequestsSVG from '@/assets/artwork/error/429-artwork.vue'
</script>

<template>
  <div class="too-many-requests-page">
    <div class="artwork">
      <TooManyRequestsSVG />
    </div>
    <h1>Maybe later.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.too-many-requests-page {
  @apply w-screen min-h-screen -mt-10;
}

.too-many-requests-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
