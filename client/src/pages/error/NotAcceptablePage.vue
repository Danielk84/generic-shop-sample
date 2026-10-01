<script setup lang="ts">
import NotAcceptableSVG from '@/assets/artwork/error/406-artwork.vue'
</script>

<template>
  <div class="not-acceptable-pag">
    <div class="artwork c-flex-all-center">
      <NotAcceptableSVG />
    </div>
    <h1>Page or content not acceptable.</h1>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.not-acceptable-pag {
  @apply w-screen min-h-screen;
}

.not-acceptable-pag .artwork {
  @apply w-full;
}

.not-acceptable-pag h1 {
  @apply text-7xl font-extrabold w-full p-10 text-center
    text-(--c-v-11);
}
</style>
