<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import icons from '@/utils/icons'
import type { ProductCardProps } from '@/components/card/types'

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))
const ImageFrameCard = defineAsyncComponent(() => import('@/components/card/ImageFrameCard.vue'))

const props = defineProps<{ data: ProductCardProps }>()
</script>

<template>
  <RouterLink :to="props.data.to" class="product-card">
    <div class="banner c-flex-all-center">
      <ImageFrameCard :img="props.data.img" :tag="props.data.tag" />
    </div>
    <div class="content">
      <div class="title">
        <span>{{ props.data.title }}</span>
      </div>
      <div class="price">
        <span>{{ props.data.price }}</span>
        <button class="add-product">
          <BaseIcon :icon="icons.card.plusCircle" size="19px"/>
        </button>
      </div>
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.product-card {
  @apply w-70 h-93 gap-10;
}

.product-card > .banner {
  @apply w-full h-90/100;
}

.product-card > .content {
  @apply py-2 px-4;
}

.product-card > .content > span {
  @apply text-nowrap overflow-hidden text-ellipsis;
}

.product-card > .content > .title {
  @apply text-2xl;
}

.product-card > .content > .price {
  @apply text-sm text-center
    flex flex-row justify-between items-center px-1;
}

.product-card > .content > .price > .add-product {
  @apply w-fit h-fit pb-1 hover:cursor-pointer;
}
</style>