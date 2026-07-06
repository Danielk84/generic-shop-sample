<script setup lang="ts">
import { defineAsyncComponent } from "vue";

import type { Tag } from "@/types/landing";
import type { Category, OfferCategory } from "@/types/api/categories";
import type { Product } from "@/types/api/products";
import type { OverviewProfile } from "@/types/api/profile";
import icons from "@/utils/icons-list";

const LandingTags = defineAsyncComponent(() => import("@/components/common/landing/LandingTags.vue"))
const LandingBanner = defineAsyncComponent(() => import("@/components/common/landing/LandingBanner.vue"))
const SectionTitle = defineAsyncComponent(() => import("@/components/common/SectionTitle.vue"))
const CategoryCard = defineAsyncComponent(() => import("@/components/common/card/CategoryCard.vue"))
const ProductVCard = defineAsyncComponent(() => import("@/components/common/card/ProductVCard.vue"))
const PriceOfferFrameCard = defineAsyncComponent(() => import("@/components/common/card/PriceOfferFrameCard.vue"))
const SellerCard = defineAsyncComponent(() => import("@/components/common/card/SellerCard.vue"))

const tags: Array<Tag> = [
  { name: "Make it you'r self", isBtn: true, icon: icons.draw_1 },
  { name: 'School & Office' },
  { name: 'Home appliances' },
  { name: 'Phone case' },
  { name: 'Cards and posters' },
  { name: 'Parties' },
  { name: 'Clothing' },
]

const categories: Array<Category> = []
const bestSales: Array<Product> = []
const specialSales: Array<OfferCategory> = []
const popularSellers: Array<OverviewProfile> = []
const newSales: Array<Product> = []
</script>

<template>
  <div class="base-home home">
    <header class="base-home">
        <LandingTags :tags="tags" />
        <LandingBanner />
    </header>
    <main>
      <!-- Categories -->
      <section v-if="categories.length !== 0" class="base-section">
        <SectionTitle :icon="icons.medalRibbonsStar">
          Categories
        </SectionTitle>
        <nav class="categories">
          <CategoryCard v-for="item of categories" :key="item.tag" :background-image="item.backgroundImage">
            {{ item.tag }}
          </CategoryCard>
        </nav>
      </section>

      <!-- Best sales -->
      <section v-if="bestSales.length !== 0" class="base-section">
        <SectionTitle :icon="icons.medalRibbonsStar" to="/">
          Best sales
        </SectionTitle>
        <nav class="best-sales">
          <ProductVCard v-for="item of bestSales" :key="item.to"
            :to="item.to" :background-image="item.backgroundImage">
            <span>{{ item.name }}</span>
            <span>{{ "Price: " + item.price }}</span>
          </ProductVCard>
        </nav>
      </section>

      <!-- Special sales -->
      <section v-if="specialSales.length !== 0" class="base-section">
        <SectionTitle :icon="icons.star">
          Special sales
        </SectionTitle>
        <nav class="special-sales">
          <PriceOfferFrameCard v-for="item of specialSales" :key="item.tag"
            to="/" :background-image="item.backgroundImage"
            :category="item.tag" :percent="item.percent" />
        </nav>
      </section>

      <!-- Popular sellers -->
      <section v-if="popularSellers.length !== 0" class="base-section">
        <SectionTitle :icon="icons.userAvatar">
          Popular sellers
        </SectionTitle>
        <nav class="popular-sellers">
          <SellerCard v-for="item of popularSellers" :key="item.username"
            to="/" :background-image="item.backgroundImage"
            :username="item.username" :total-sells="item.totalSells" :followers="item.followers"/>
        </nav>
      </section>

      <!-- New products -->
      <section v-if="newSales.length !== 0" class="base-section">
        <SectionTitle :icon="icons.newProduct" to="/">
          New sales
        </SectionTitle>
        <nav class="best-sales">
          <ProductVCard v-for="item of newSales" :key="item.to"
            :to="item.to" :background-image="item.backgroundImage">
            <span>{{ item.name }}</span>
            <span>{{ "Price: " + item.price }}</span>
          </ProductVCard>
        </nav>
      </section>
    </main>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.base-home {
  @apply flex flex-col items-center gap-7 w-full;
}

.base-home main {
  @apply flex flex-col items-center gap-10
}

.home {
  @apply min-h-140;
}

.base-section {
  @apply flex flex-col px-10 gap-10 w-full;
}

.base-section .categories {
  @apply flex flex-row items-center justify-center gap-7;
}

.base-section .best-sales {
  @apply flex flex-row flex-wrap items-center justify-center gap-7;
}

.base-section .popular-sellers {
  @apply flex flex-row flex-wrap items-center justify-center gap-7;
}

.base-section .special-sales {
  @apply flex flex-row items-center justify-center gap-10;
}
</style>
