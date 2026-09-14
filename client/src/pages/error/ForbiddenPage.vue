<script setup lang="ts">
import ForbiddenSVG from '@/assets/artwork/error/403-artwork.vue'
</script>

<template>
  <div class="forbidden-page">
    <div class="artwork">
      <ForbiddenSVG />
    </div>
    <h1>Your Access are Forbidden.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.forbidden-page {
  @apply w-screen min-h-screen -mt-30;
}

.forbidden-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg;
}
</style>
