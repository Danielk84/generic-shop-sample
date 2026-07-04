<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import icons from '@/utils/icons-list';

const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))
</script>

<template>
  <div class="landing-banner">
    <div class="message">
      <h1>Wellcome to our generic shop</h1>
      <p>
        There is a lot of different item in generic shop so you can buy items and sell your
        products. Wow!!
      </p>
    </div>
    <div class="btn-box">
      <RouterLink to="/">
        <div class="base-btn btn-1">
          <span>See all products</span>
        </div>
      </RouterLink>
      <RouterLink to="/">
        <div class="base-btn btn-2">
          <span>Sell you'r products</span>
          <BaseIcon :icon="icons.draw_2" fill-color="--color-btn-text-1"/>
        </div>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.landing-banner {
  @apply flex flex-row items-center justify-evenly gap-5
  bg-landing-banner-bg
    bg-[url(@/assets/images/landing-banner.png)]
    rounded-2xl w-9/12 min-h-120;
}

.landing-banner .message {
  @apply text-pretty flex flex-col gap-10 justify-center items-center
    max-w-120 text-btn-text-1;
}

.landing-banner .message h1 {
  @apply text-5xl font-bold;
}

.landing-banner .message p {
  @apply text-xl italic;
}

.landing-banner .btn-box {
  @apply flex flex-col items-center justify-center gap-10;
}

.landing-banner .base-btn {
  @apply flex flex-row gap-1 items-center justify-center p-4 rounded-xl
    outline-2 w-60 font-bold
    transition delay-250 hover:ring-4 hover:brightness-125;
}

.landing-banner .btn-1 {
  @apply outline-btn-text-1 text-btn-text-1
    ring-btn-text-1;
}

.landing-banner .btn-2 {
  @apply outline-primary-color bg-primary-color text-btn-text-1
    ring-primary-color;
}
</style>
