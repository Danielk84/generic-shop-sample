<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

const NavBar = defineAsyncComponent(() => import('@/components/common/nav-bar/NavBar.vue'))
const FooterBlock = defineAsyncComponent(() => import('@/components/common/footer/FooterBlock.vue'))
</script>

<template>
  <div class="base">
    <header>
      <NavBar />
    </header>
    <div class="page">
      <main>
        <RouterView />
      </main>
      <footer>
        <FooterBlock />
      </footer>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base > .page {
  @apply pt-22 max-w-screen min-h-screen flex flex-col gap-30;
}

.base > .page > main {
  @apply w-screen h-full relative;
}

.base > .page > footer {
  @apply w-screen h-full z-40 relative;
}
</style>
