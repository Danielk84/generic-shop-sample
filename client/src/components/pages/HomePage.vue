<script setup lang="ts">
import { defineAsyncComponent } from "vue";

const LandingTags = defineAsyncComponent(() => import("@/components/common/landing/LandingTags.vue"))
const LandingBanner = defineAsyncComponent(() => import("@/components/common/landing/LandingBanner.vue"))
const SectionTitle = defineAsyncComponent(() => import("@/components/common/SectionTitle.vue"))
const CategoryCard = defineAsyncComponent(() => import("@/components/common/card/CategoryCard.vue"))
const ProductVCard = defineAsyncComponent(() => import("@/components/common/card/ProductVCard.vue"))
const PriceOfferFrameCard = defineAsyncComponent(() => import("@/components/common/card/PriceOfferFrameCard.vue"))
const SellerCard = defineAsyncComponent(() => import("@/components/common/card/SellerCard.vue"))
</script>

<template>
  <div class="base-home home">
    <header class="base-home">
        <LandingTags />
        <LandingBanner />
    </header>
    <main>
      <!-- Categories -->
      <section class="base-section">
        <SectionTitle
          msg="Categories" />
        <nav class="categories">
          <CategoryCard />
          <CategoryCard />
          <CategoryCard />
          <CategoryCard />
          <CategoryCard />
          <CategoryCard />
        </nav>
      </section>

      <!-- Best sales -->
      <section class="base-section">
        <SectionTitle
          msg="Best sales"
          see-also-link="/" />
        <nav class="best-sales">
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
        </nav>
      </section>

      <!-- Special sales -->
      <section class="base-section">
        <SectionTitle
        msg="Special sales" />
        <nav class="special-sales">
          <PriceOfferFrameCard to="/" category="Spetial" percent=20 />
          <PriceOfferFrameCard to="/" category="Spetial" percent=20 />
          <PriceOfferFrameCard to="/" category="Spetial" percent=20 />
          <PriceOfferFrameCard to="/" category="Spetial" percent=20 />
          <PriceOfferFrameCard to="/" category="Spetial" percent=20 />
        </nav>
      </section>

      <!-- Popular sellers -->
      <section class="base-section">
        <SectionTitle
          msg="Popular sellers" />
        <nav class="popular-sellers">
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
          <SellerCard name="dk" to="/" total-sells="400" followers="240"/>
        </nav>
      </section>

      <!-- New products -->
      <section class="base-section">
        <SectionTitle
          msg="New products"
          see-also-link="/" />
        <nav class="best-sales">
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
          <ProductVCard to="/" name="Products" price="100.00"/>
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
