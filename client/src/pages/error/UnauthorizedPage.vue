<script setup lang="ts">
import UnauthorizedSVG from '@/assets/artwork/error/401-artwork.vue'
</script>

<template>
  <div class="unauthorized-page">
    <div class="artwork">
      <UnauthorizedSVG />
    </div>
    <RouterLink :to="{ name: 'auth' }">
      <h1>Authenticate first.</h1>
    </RouterLink>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.unauthorized-page {
  @apply w-screen min-h-screen -mt-20;
}

.unauthorized-page h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-error-p-msg underline;
}
</style>
