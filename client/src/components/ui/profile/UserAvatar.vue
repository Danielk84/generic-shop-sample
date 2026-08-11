<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

const ImageFrameCard = defineAsyncComponent(() => import('@/components/card/ImageFrameCard.vue'))

const props = defineProps<{ img?: string }>()
</script>

<template>
  <div class="user-avatar">
    <ImageFrameCard :img="props.img" />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.user-avatar {
  @apply size-full aspect-square rounded-full overflow-hidden;
}
</style>