<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const SearchBox = defineAsyncComponent(() => import('@/components/ui/search/SearchBox.vue'))
</script>

<template>
  <div class="landing-page">
    <div class="landing-banner c-flex-all-center">
      <div class="banner-content c-flex-all-center">
        <h1 class="c-flex-all-center">
          <span>
            Crafting Comfort, Redifining Spaces.
          </span>
          <span>
            Your Home, Your Signature Style.
          </span>
        </h1>
        <p>
          Lorem, ipsum dolor sit amet consectetur adipisicing elit. Exercitationem ipsam velit temporibus animi natus, autem molestias aperiam dolorum! Sint excepturi molestias nemo eligendi ducimus autem atque quo magni voluptatum cupiditate.
        </p>
      </div>
      <div class="banner-search">
        <SearchBox />
      </div>
    </div>
    <div class="landing-content c-flex-all-center">
      
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.landing-page {
  @apply w-screen h-full relative -top-25
    bg-landing-p-primary-bg;
}

.landing-page > .landing-banner {
  @apply pt-32 w-full h-200
    flex-col gap-20;
}

.landing-page > .landing-banner > .banner-content {
  @apply text-landing-p-text w-50/100 flex-col gap-5;
}

.landing-page > .landing-banner > .banner-content > h1 {
  @apply text-5xl flex-col;
}

.landing-page > .landing-banner > .banner-content > p {
  @apply w-80/100;
}

.landing-page > .landing-banner > .banner-search {
  @apply w-100 h-fit;
} 

.landing-page > .landing-content {
  @apply w-full h-fit z-10
    rounded-t-4xl pt-20 gap-10
    bg-landing-p-secondary-bg;
}
</style>