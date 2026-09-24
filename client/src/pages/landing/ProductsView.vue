<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'

import icons from '@/utils/icons'
import { useTimer } from '@/utils/helper'
import type { ProductSummaryResponse } from '@/contracts/products/response.interface'

const ProductCard = defineAsyncComponent(
  () => import('@/components/card/ProductCard.vue'),
)
const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const props = defineProps<{
  title: string
  to: string
  emptyView: string
  data?: ProductSummaryResponse[]
}>()

const scrollBoxRef = ref<HTMLDivElement | null>(null)
const isDragging = ref(false)
const startX = ref(0)
const initialScrollLeft = ref(0)

function onPointerDown(event: PointerEvent) {
  const element = scrollBoxRef.value
  if (!element) return

  isDragging.value = true
  startX.value = event.clientX
  initialScrollLeft.value = element.scrollLeft

  element.setPointerCapture(event.pointerId)
}

function onPointerMove(event: PointerEvent) {
  const element = scrollBoxRef.value
  if (!element || !isDragging.value) return

  const distance = event.clientX - startX.value
  element.scrollLeft = initialScrollLeft.value - distance
}

function onPointerUp(event: PointerEvent) {
  const element = scrollBoxRef.value

  isDragging.value = false

  if (element?.hasPointerCapture(event.pointerId)) {
    element.releasePointerCapture(event.pointerId)
  }
}

const scrollStep = 15
const scrollTimeout = 24
const goForward = useTimer(() => {
  const element = scrollBoxRef.value
  if (!element) return

  if (initialScrollLeft.value <= element.scrollWidth) {
    initialScrollLeft.value += scrollStep
    element.scrollLeft = initialScrollLeft.value
  }
}, scrollTimeout)
</script>

<template>
  <div
    v-if="props.data === undefined || props.data.length === 0"
    class="empty-view c-flex-all-center"
  >
    <h2>{{ props.emptyView }}</h2>
  </div>
  <div v-else class="products-view">
    <div class="label c-flex-all-center">
      <h2>New Products</h2>
      <div class="line"></div>
      <RouterLink class="link" to="/"> See more... </RouterLink>
    </div>
    <div
      ref="scrollBoxRef"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerUp"
      class="products"
      :class="{ 'is-draging': isDragging }"
    >
      <ProductCard
        class="product-item"
        v-for="item of props.data"
        :key="item.id"
        :data="{
          to: `/products/${item.id}`,
          title: item.name,
          price: item.price,
          img: item.img_path,
          alt: `${item.name}-${item.pub_date}`,
        }"
      />
    </div>
    <div class="moving-box forward-box">
      <button
        @pointerdown="goForward.start()"
        @pointerup="goForward.end()"
        @touchstart="goForward.start()"
        @touchend="goForward.end()"
        class="moving-btn"
      >
        <BaseIcon :icon="icons.pages.landing.rightArrow" />
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.empty-view {
  @apply font-bold text-4xl w-full h-fit p-20;
}

.products-view {
  @apply w-full h-fit px-10
    relative isolate;
}

.products-view .label {
  @apply gap-5;
}

.label > h2 {
  @apply font-bold text-4xl text-nowrap m-10;
}

.label > .line {
  @apply w-full h-fit border-2
    border-products-view-line;
}

.label > .link {
  @apply text-2xl font-bold text-nowrap
    text-products-view-link;
}

.products-view .products {
  scrollbar-width: none;

  @apply flex flex-row items-start
    gap-10 p-10 w-full h-fit
    overflow-x-hidden flex-nowrap
    scroll-pl-20 touch-pan-x
    transition duration-600 ease-in;
}

.products-view .is-draging {
  @apply cursor-grabbing;
}

.products-view .products::-webkit-scrollbar {
  display: none;
}

.products .product-item {
  @apply shrink-0;
}

.products-view .moving-box {
  @apply w-fit p-2 h-full absolute z-10
    flex items-center justify-center;
}

.moving-box .moving-btn {
  @apply p-3 rounded-full
    border
    cursor-pointer bg-white;
}

.products-view .forward-box {
  @apply right-9 top-0;
}
</style>
