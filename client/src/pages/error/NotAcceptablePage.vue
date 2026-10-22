<script setup lang="ts">
import NotAcceptableSVG from '@/assets/artwork/error/406-artwork.vue'
</script>

<template>
  <div class="not-acceptable-pag c-flex-all-center">
    <div class="artwork c-flex-all-center">
      <NotAcceptableSVG />
    </div>
    <h1>Page or content not acceptable.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.not-acceptable-pag {
  @apply w-full min-h-screen flex-col md:mt-5;
}

.not-acceptable-pag .artwork {
  @apply w-60/100 max-h-100;
}

.not-acceptable-pag h1 {
  @apply text-4xl md:text-7xl font-extrabold w-full p-6 md:p-10 text-center
    text-(--c-v-11);
}
</style>
