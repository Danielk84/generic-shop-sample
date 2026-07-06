<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import icons from '@/utils/icons-list'

const BaseBtn = defineAsyncComponent(() => import('@/components/common/button/BaseBtn.vue'))
const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
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
      <RouterLink to="/" class="base-btn-size">
        <BaseBtn :click="() => {}" :style="{
          backgroundColor: 'none',
          outlineColor: '--color-secondary-text',
          outlineWidth: '2px',
          color: '--color-secondary-text', 
        }">
          <span>See all products</span>
        </BaseBtn>
      </RouterLink>
      <RouterLink to="/" class="base-btn-size">
        <BaseBtn>
          <span>Sell you'r products</span>
          <BaseIcon :icon="icons.draw_2" fill-color="--color-secondary-text" />
        </BaseBtn>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.landing-banner {
  @apply flex flex-row items-center justify-evenly gap-5
    bg-default-banner
    bg-[url(@/assets/images/landing-banner.png)]
    rounded-2xl w-9/12 min-h-120;
}

.landing-banner .message {
  @apply text-pretty flex flex-col gap-10 justify-center items-center
    max-w-120 text-secondary-text;
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
</style>
