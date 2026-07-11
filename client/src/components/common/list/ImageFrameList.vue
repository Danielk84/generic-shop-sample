<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue';

import type { ImageFrame } from '@/types/cards';
import icons from "@/utils/icons-list"

const ImageFrameCard = defineAsyncComponent(() => import("@/components/common/card/ImageFrameCard.vue"))
const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))
const FullScreenImageFrame = defineAsyncComponent(() => import("@/components/common/FullScreenImageFrame.vue"))

const props = defineProps<{ imgList: Array<ImageFrame> }>()

const showUp = ref<ImageFrame>(props.imgList.length !== 0?
  props.imgList[0] as ImageFrame :
  { src: "", alt: "empty!" }
)

const fullScreen = ref<boolean>(false)
</script>

<template>
  <div class="image-frame-list">
    <div class="show-up">
      <ImageFrameCard :src="showUp.src" :alt="showUp?.alt">
        <button class="full-screen-btn" @click="fullScreen = true">
          <BaseIcon :icon="icons.fullScreen" fill-color="--color-secondary-theme"/>
        </button>
      </ImageFrameCard>
    </div>
    <div class="img-list default-scrollbar">
      <div class="item" v-for="item of props.imgList" :key="item.src">
        <ImageFrameCard @click="() => { showUp = item }"
          :src="item.src"
          :alt="item.alt"/>
      </div>
    </div>
    <div v-if="fullScreen">
      <FullScreenImageFrame
        :src="showUp.src"
        :alt="showUp.alt"
        @destroy="fullScreen = false" />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.image-frame-list {
  @apply flex-col w-full h-full gap-2;
}

.image-frame-list .show-up {
  @apply w-full h-70/100 hover:outline rounded-lg outline-baseline cursor-pointer;
}

.image-frame-list .show-up .full-screen-btn {
  @apply w-full h-full flex justify-end items-end p-4;
}

.image-frame-list .img-list {
  @apply min-w-0 w-full py-3 px-1
    flex items-center flex-row flex-nowrap gap-4
    overflow-x-scroll overflow-y-hidden;
}

.image-frame-list .img-list .item {
  @apply size-30 shrink-0 cursor-pointer;
}
</style>