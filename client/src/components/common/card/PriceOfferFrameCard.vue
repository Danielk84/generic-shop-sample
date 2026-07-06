<script setup lang="ts">
import { defineAsyncComponent, computed, type StyleValue } from 'vue'

import type { PriceOfferFrame } from '@/types/cards';
import { getCssUrl } from '@/utils/css-helper';

const OfferBookmarkSVG = defineAsyncComponent(() => import('@/assets/images/offer-bookmark.svg'))

const props = defineProps<PriceOfferFrame>()

const style = computed<StyleValue>(() => ({
  backgroundImage: getCssUrl(props.backgroundImage),
}))
</script>

<template>
  <RouterLink :to="props.to" class="price-offer-frame default-banner-bg" :style="style">
    <OfferBookmarkSVG class="offer-bookmark" />
    <div class="percent c-flex-all-center">
      <span>{{ props.percent + '%' }}</span>
      <span>off</span>
    </div>
    <div class="content">
      <span>{{ props.category }}</span>
    </div>
  </RouterLink>
</template>

<style scoped>
@reference "@/styles/index.css";

.price-offer-frame {
  @apply rounded-2xl w-60 h-60;
}

.price-offer-frame .offer-bookmark {
  @apply z-0 w-32 h-43 relative -top-8 left-26
    stroke-secondary-theme fill-primary-theme;
}

.price-offer-frame .percent {
  @apply flex-col
    z-10 text-xl font-bold relative bottom-40 left-12.5
    text-secondary-text;
}

.price-offer-frame .content {
  @apply z-10 w-50 relative bottom-12 left-5 text-2xl
    text-secondary-text font-bold text-ellipsis;
}
</style>
