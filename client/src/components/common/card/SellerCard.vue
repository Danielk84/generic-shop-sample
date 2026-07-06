<script setup lang="ts">
import { defineAsyncComponent, computed, type StyleValue } from 'vue'

import type { Seller } from '@/types/cards';
import icons from '@/utils/icons-list'
import { getCssUrl } from '@/utils/css-helper';

const BaseIcon = defineAsyncComponent(() => import('@/components/common/BaseIcon.vue'))
const BaseBtn = defineAsyncComponent(() => import("@/components/common/button/BaseBtn.vue"))

const props = defineProps<Seller>()

const style = computed<StyleValue>(() => ({
  backgroundImage: getCssUrl(props.backgroundImage),
}))
</script>

<template>
  <RouterLink :to="props.to" class="seller-card default-shadow-hover">
    <div class="top-box">
      <div class="top-banner default-banner-bg">
        <div v-if="!props.backgroundImage" class="profile c-flex-all-center">
          <BaseIcon :icon="icons.userAvatar" size="80px" stroke-color="--color-secondary-theme"/>
        </div>
        <div v-else class="profile c-flex-all-center" :style="style"></div>
      </div>
    </div>
    <div class="down-box c-flex-all-center">
      <div class="username c-flex-all-center">
        <span>{{ props.username }}</span>
      </div>
      <div class="info">
        <span class="border-r border-baseline px-1"> total sells: {{ props.totalSells}} </span>
        <span> followers: {{ props.followers }} </span>
      </div>
      <div class="btn">
        <BaseBtn :style="{
          backgroundColor: 'none',
          outlineColor: '--color-primary-theme',
          color: '--color-primary-theme',
        }">
          <span>Follow</span>
          <BaseIcon
            :icon="icons.userFollow"
            fill-color="--color-primary-theme"
            stroke-color="--color-primary-theme"
          />
        </BaseBtn>
      </div>
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.seller-card {
  @apply flex flex-col items-center justify-between gap-6 rounded-2xl
    border border-baseline
    w-95 h-145;
}

.seller-card .top-box {
  @apply w-full h-50/100;
}

.seller-card .down-box {
  @apply flex-col gap-6 my-2 w-full;
}

.seller-card .top-banner {
  @apply w-full h-45/100
  border-b-4 border-b-primary-theme
  rounded-t-2xl static;
}

.seller-card .profile {
  @apply rounded-full
  size-28 bg-default-banner
  border-5 border-primary-theme border-t-0 border-b-0
  relative top-19 left-33;
}

.seller-card .username {
  @apply text-2xl font-bold text-nowrap;
}

.seller-card .btn {
  @apply w-93/100 h-17 py-2;
}
</style>
