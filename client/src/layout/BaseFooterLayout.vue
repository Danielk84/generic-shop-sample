<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const BaseFooter = defineAsyncComponent(
  () => import('@/components/common/footer/BaseFooter.vue'),
)
</script>

<template>
  <div class="page">
    <main>
      <RouterView />
    </main>
    <footer>
      <BaseFooter />
    </footer>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.page {
  @apply pt-22 max-w-screen min-h-screen flex-col;
}
</style>
