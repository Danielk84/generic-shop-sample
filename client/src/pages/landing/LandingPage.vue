<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const SearchBox = defineAsyncComponent(
  () => import('@/components/ui/search/SearchBox.vue'),
)
const LatestProducts = defineAsyncComponent(
  () => import('@/pages/landing/LastestProducts.vue'),
)
const PopularProducts = defineAsyncComponent(
  () => import('@/pages/landing/PopularProducts.vue'),
)
</script>

<template>
  <div class="landing-page">
    <div class="landing-banner c-flex-all-center">
      <div class="banner-content c-flex-all-center">
        <h1 class="c-flex-all-center">
          <span> Crafting Comfort, Redifining Spaces. </span>
          <span> Discover Things You'll Love. </span>
        </h1>
        <p>
          Shop thoughtfully selected products designed to make everyday life
          better. From everyday essentials to unique finds, discover quality
          products, simple shopping, and a seamless experience all in one place.
          Our most loved products, chosen by shoppers like you.
        </p>
      </div>
      <div class="banner-search">
        <SearchBox />
      </div>
    </div>
    <div class="landing-content c-flex-all-center">
      <LatestProducts />
      <PopularProducts />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.landing-page {
  @apply w-full min-h-screen relative -top-25
    bg-radial
    from-(--c-v-3) from-0%
    to-(--c-v-1) to-100%;
}

.landing-page > .landing-banner {
  @apply pt-32 w-full h-auto min-h-200
    flex-col gap-10 md:gap-20
    px-4 md:px-0;
}

.landing-page > .landing-banner > .banner-content {
  @apply text-(--c-v-6) w-full md:w-50/100 flex-col gap-6 md:gap-10
    items-center text-center;
}

.landing-page > .landing-banner > .banner-content > h1 {
  @apply text-3xl md:text-5xl font-extrabold flex-col font-secondary-default
    w-full h-full leading-12 md:leading-16;
}

.landing-page > .landing-banner > .banner-content > p {
  @apply w-full md:w-80/100 text-base md:text-lg;
}

.landing-page > .landing-banner > .banner-search {
  @apply w-full md:w-100 h-fit;
}

.landing-page > .landing-content {
  @apply w-full h-fit z-10 flex-col
    rounded-t-4xl pt-10 md:pt-20 gap-6 md:gap-10 px-4 md:px-0
    bg-(--c-v-7) dark:bg-(--c-v-0);
}
</style>
