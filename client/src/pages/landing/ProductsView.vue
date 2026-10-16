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

// Manual drag scrolling is only handled for mouse pointers. Touch and pen input
// keep the native horizontal scroll so that taps still navigate the cards.
const DRAG_THRESHOLD = 6

const isPointerDown = ref(false)
const isDragging = ref(false)
const activePointerId = ref<number | null>(null)
const startX = ref(0)
const startScrollLeft = ref(0)

function onPointerDown(event: PointerEvent) {
  if (event.pointerType !== 'mouse') return

  const element = scrollBoxRef.value
  if (!element) return

  isPointerDown.value = true
  isDragging.value = false
  activePointerId.value = event.pointerId
  startX.value = event.clientX
  startScrollLeft.value = element.scrollLeft
}

function onPointerMove(event: PointerEvent) {
  const element = scrollBoxRef.value
  if (
    !element ||
    !isPointerDown.value ||
    event.pointerId !== activePointerId.value
  ) {
    return
  }

  const distance = event.clientX - startX.value
  if (!isDragging.value && Math.abs(distance) < DRAG_THRESHOLD) return

  // Capture only once the drag really starts, otherwise capturing on every
  // pointerdown swallows the click event and breaks card navigation.
  if (!isDragging.value) {
    isDragging.value = true
    element.setPointerCapture(event.pointerId)
  }

  element.scrollLeft = startScrollLeft.value - distance
}

function endPointer(event: PointerEvent) {
  const element = scrollBoxRef.value

  if (
    activePointerId.value !== null &&
    event.pointerId !== activePointerId.value
  ) {
    return
  }

  if (
    element &&
    isDragging.value &&
    element.hasPointerCapture(event.pointerId)
  ) {
    element.releasePointerCapture(event.pointerId)
  }

  isPointerDown.value = false
  isDragging.value = false
  activePointerId.value = null
}

// Resets a pending press that left the list before the drag threshold was
// reached. An active drag is left alone so capturing cannot cut it short.
function onPointerLeave(event: PointerEvent) {
  if (isDragging.value) return
  if (
    activePointerId.value !== null &&
    event.pointerId !== activePointerId.value
  ) {
    return
  }

  isPointerDown.value = false
  activePointerId.value = null
}

const scrollStep = 12
const scrollTimeout = 16
const goForward = useTimer(() => {
  const element = scrollBoxRef.value
  if (!element) return

  const maxScroll = element.scrollWidth - element.clientWidth
  element.scrollLeft = Math.min(element.scrollLeft + scrollStep, maxScroll)
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
      <h2>{{ props.title }}</h2>
      <div class="line"></div>
      <RouterLink class="link" :to="props.to"> See more... </RouterLink>
    </div>
    <div
      ref="scrollBoxRef"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="endPointer"
      @pointercancel="endPointer"
      @pointerleave="onPointerLeave"
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
        @pointercancel="goForward.end()"
        @pointerleave="goForward.end()"
        @touchstart="goForward.start()"
        @touchend="goForward.end()"
        @touchcancel="goForward.end()"
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
  @apply font-bold text-2xl md:text-4xl w-full h-fit p-10 md:p-20;
}

.products-view {
  @apply w-full h-fit px-4 md:px-10
    relative isolate;
}

.products-view .label {
  @apply gap-3 md:gap-5 flex flex-row items-center justify-between;
}

.label > h2 {
  @apply font-bold text-2xl md:text-4xl text-nowrap m-4 md:m-10;
}

.label > .line {
  @apply w-full h-fit border-2
    border-(--c-v-4) dark:border-(--c-v-7);
}

.label > .link {
  @apply text-lg md:text-2xl font-bold text-nowrap
    text-(--c-v-11);
}

.products-view .products {
  scrollbar-width: none;

  @apply flex flex-row items-start
    gap-6 md:gap-10 p-4 md:p-10 w-full h-fit
    overflow-x-auto flex-nowrap
    scroll-pl-20 touch-pan-x
    transition duration-600 ease-in;
}

.products-view .is-draging {
  @apply cursor-grabbing select-none;
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
    cursor-pointer
    bg-(--c-v-7) dark:bg-(--c-v-2);
}

.products-view .forward-box {
  @apply right-4 md:right-9 top-0;
}
</style>
