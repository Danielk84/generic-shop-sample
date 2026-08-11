<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

const SubscribeBox = defineAsyncComponent(() => import('@/components/common/footer/SubscribeBox.vue'))
const BaseFooter = defineAsyncComponent(() => import('@/components/common/footer/BaseFooter.vue'))
</script>

<template>
  <div class="footer-block">
    <div class="footer-content c-flex-all-center">
      <h4 class="c-flex-all-center">
        <span>Subscribe To Your Newsletter To Stay</span>
        <span>Updated About Discounts</span>
      </h4>
      <div class="footer-sub">
        <SubscribeBox />
      </div>
    </div>
    <BaseFooter />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.footer-block {
  @apply w-full h-fit;
}

.footer-block > .footer-content {
  @apply bg-footer-bg text-footer-text p-10
    flex-col gap-10 text-4xl;
}

.footer-block > .footer-content > h4 {
  @apply flex-col;
}

.footer-block > .footer-content > .footer-sub {
  @apply w-90 h-fit;
}
</style>