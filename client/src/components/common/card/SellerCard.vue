<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import icons from '@/utils/icons-list';

const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))

const props = defineProps<{
  to: string,
  name: string,
  totalSells: string,
  followers: string,
}>()
</script>

<template>
  <RouterLink :to="props.to" class="seller-card">
    <div class="top-box">
      <div class="top-banner">
        <div class="profile">
          <BaseIcon :icon="icons.userAvatar" />
        </div>
      </div>
    </div>
    <div class="username">
      <span>{{ props.name }}</span>
    </div>
    <div class="info">
      <span class="border-r border-navigation-bar px-1">
        total sells: {{ props.totalSells }}
      </span>
      <span>
        followers: {{ props.followers }}
      </span>
    </div>
    <div class="follow-btn">
      <span>Follow</span>
      <BaseIcon
        :icon="icons.userFollow"
        fill-color="--color-primary-color"
        stroke-color="--color-primary-color" />
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.seller-card {
  @apply flex flex-col items-center justify-between gap-6 rounded-2xl
    border border-navigation-bar
    hover:shadow-[0px_0px_8px] shadow-navigation-bar
    w-95 h-145;
    
  }
  
  .seller-card .top-box {
    @apply w-full h-50/100;
  }
  
  .seller-card .top-banner {
    @apply w-full h-45/100 bg-polar-night-2
    rounded-t-2xl static;
  }
  
  .seller-card .profile {
    @apply flex items-center justify-center rounded-full
    size-28 bg-navigation-bar
    relative top-19 left-33;
}

.seller-card .username {
  @apply flex items-center justify-center
    text-2xl font-bold text-nowrap;
}

.seller-card .follow-btn {
  @apply flex flex-row items-center justify-center gap-2
    h-12 w-90/100 border-2 rounded-lg m-4
    border-primary-color text-primary-color;
}
</style>