<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue';

import type { ImageFrameCardProps } from '@/components/card/types';

const ImageFrameList = defineAsyncComponent(() => import('@/pages/products/ImageFrameList.vue'))

const images = ref<ImageFrameCardProps[]>([
  { img: "/1.png", alt:"product images" },
  { img: "/3333.png" },
  { img: "/2.png", alt:"product images" },
  { img: "/3.png", alt:"product images" },
  { img: "/3.png", alt:"product images" },
  { img: "/4.png", alt:"product images" },
  { img: "/5.png", alt:"product images" },
  { img: "/6.png", alt:"product images" },
  { img: "/7.png", alt:"product images" },
  { img: "/8.png", alt:"product images" },
])
</script>

<template>
  <div class="w-screen h-500">
    <ImageFrameList :data="{ images: images }" />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>