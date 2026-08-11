<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import type { BlogCardProps } from '@/components/card/types'

const ImageFrameCard = defineAsyncComponent(() => import('@/components/card/ImageFrameCard.vue'))
const UserAvatar = defineAsyncComponent(() => import('@/components/ui/profile/UserAvatar.vue'))

const props = defineProps<{ data: BlogCardProps }>()
</script>

<template>
  <RouterLink :to="props.data.to" class="blog-card c-flex-all-center">
    <div class="banner c-flex-all-center">
      <ImageFrameCard :img="props.data.banner" />
    </div>
    <div class="blog-content">
      <div class="info">
        <div class="blog-user-avatar">
          <UserAvatar :img="props.data.profileImage" />
        </div>
        <span>{{ props.data.username }}</span>
        <span class="separator">&middot;</span>
        <span>{{ props.data.pudDate }}</span>
      </div>
      <div class="title">
        <h2>{{ props.data.title }}</h2>
      </div>
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.blog-card {
  @apply w-120 h-120 flex-col;
}

.blog-card > .banner {
  @apply w-98/100 h-60/100;
}

.blog-card > .blog-content {
  @apply flex flex-col justify-start items-start px-4 py-1;
}

.blog-card > .blog-content > .info {
  @apply flex flex-row justify-start items-center
    gap-5 text-lg font-bold
    w-full h-15;
}

.blog-card > .blog-content > .info > .blog-user-avatar {
  @apply size-13;
}

.blog-card > .blog-content > .info > .separator {
  @apply text-7xl;
}

.blog-card > .blog-content > .title {
  @apply text-wrap overflow-hidden text-2xl text-ellipsis;
}
</style>