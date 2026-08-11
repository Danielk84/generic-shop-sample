<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue';

import icons from '@/utils/icons'
import type { OrderItemsCardProps } from '@/pages/basket/types'

const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))
const ImageFrameCard = defineAsyncComponent(() => import('@/components/card/ImageFrameCard.vue'))

const props = defineProps<{ data: OrderItemsCardProps }>()
const emits = defineEmits<{
  (e: 'changeCount', count: number): void
  (e: 'destroy'): void
}>()

const count = ref<number>(props.data.count)
const isSaved = ref<boolean>(true)

function saveChange() {
  emits('changeCount', count.value)
}
</script>

<template>
  <div class="order-item">
    <button class="remove-btn c-flex-all-center">
      <BaseIcon :icon="icons.pages.basket.close"/>
    </button>
    <div class="order-info">
      <div class="img">
        <ImageFrameCard :img="props.data.img" :alt="props.data.alt" />
      </div>
      <div class="item c-flex-all-center">
        <div class="sub-item sub-1">
          <div class="name">
            <h2>{{ props.data.name }}</h2>
          </div>
          <div class="info">
            <h3>price: {{ props.data.price }}</h3>
            <h3>total: {{ props.data.total }}</h3>
          </div>
        </div>
        <div class="sub-item sub-2">
          <div class="change-box c-flex-all-center">
            <button
              class="change-btn change-box-item"
              @click="() => {
                count -= 1;
                isSaved = false
              }"
            >
              <BaseIcon
                :icon="icons.pages.basket.minus"
                stroke-color="--color-basket-item-change-icon"
                fill-color="--color-basket-item-change-icon"
              />
            </button>
            <span class="change-box-item">{{ count }}</span>
            <button
              class="change-btn change-box-item"
              @click="() => {
                count += 1;
                isSaved = false
              }"
            >
              <BaseIcon
                :icon="icons.pages.basket.plus"
                stroke-color="--color-basket-item-change-icon"
                fill-color="--color-basket-item-change-icon"
              />
            </button>
          </div>
          <button
            class="save-btn"
            :class="{ 'not-saved': !isSaved }"
            @click="saveChange()"
          >
            <span>{{ isSaved ? 'saved' : 'save changes' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.order-item {
  @apply w-150 h-30 flex flex-row justify-between
    hover:backdrop-brightness-90;
}

.order-item > .remove-btn {
  @apply cursor-pointer w-7/100 h-full
    hover:bg-basket-item-remove-hover/10;
}

.order-item > .order-info {
  @apply flex flex-row items-center justify-between
    w-91/100 h-full;
}

.order-item > .order-info > .img {
  @apply size-19;
}

.order-item > .order-info > .item {
  @apply flex-row gap-2 size-full;
}

.order-info .sub-item {
  @apply flex flex-col justify-center
    h-full gap-2;
}

.order-info .sub-1 {
  @apply px-2 w-60/100;
}

.order-info .sub-2 {
  @apply items-center w-40/100;
}

.sub-item .name {
  @apply h-fit w-fit;
}

.order-info h2 {
  @apply text-xl text-nowrap text-ellipsis;
}

.order-info .change-box {
  @apply flex-row w-fit h-fit
    rounded-xl border overflow-hidden
    border-basket-item-change-border;
}

.order-info .change-box-item {
  @apply flex justify-center items-center
    size-10;
}

.order-item .change-btn {
  @apply cursor-pointer
    bg-basket-item-change-btn;
}

.sub-item .info {
  @apply flex flex-row gap-4;
} 

.sub-item h3 {
  @apply w-fit h-fit text-nowrap;
}

.sub-item .save-btn {
  @apply w-65/100 h-10 border rounded-xl font-bold
    bg-none;
}

.sub-item .not-saved {
  @apply bg-basket-item-not-saved-btn
    text-basket-item-not-saved-text
    cursor-pointer;
}
</style>