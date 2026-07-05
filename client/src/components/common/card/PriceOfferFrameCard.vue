<script setup lang="ts">
import { defineAsyncComponent } from "vue";

const OfferBookmarkSVG = defineAsyncComponent(() => import("@/assets/images/offer-bookmark.svg"))

const props = defineProps<{
  to: string,
  category: string,
  percent: string,
}>()

</script>

<template>
  <RouterLink :to="props.to" class="price-offer-frame">
    <OfferBookmarkSVG class="offer-bookmark" />
    <div class="percent">
      <span>{{ props.percent + "%" }}</span>
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
  @apply rounded-2xl w-60 h-60 bg-landing-banner-bg;
}

.price-offer-frame .offer-bookmark {
  @apply z-0 w-32 h-43 relative -top-8 left-26
    stroke-secondary-color fill-primary-color;
}

.price-offer-frame .percent {
  @apply flex flex-col justify-center items-center
    z-10 text-xl font-bold relative bottom-40 left-12.5
    text-btn-text-1;
}

.price-offer-frame .content {
  @apply z-10 w-50 relative bottom-12 left-5 text-2xl
    text-btn-text-1 font-bold text-ellipsis;
}
</style>