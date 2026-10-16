<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'

import icons from '@/utils/icons'
import type { OrderItemsCardProps } from '@/pages/basket/types'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
const ImageFrameCard = defineAsyncComponent(
  () => import('@/components/card/ImageFrameCard.vue'),
)

const props = defineProps<{ data: OrderItemsCardProps }>()
const emits = defineEmits<{
  (e: 'changeCount', count: number): void
  (e: 'destroy'): void
}>()

const count = ref<number>(props.data.count)
const hasChanges = ref<boolean>(false)

watch(count, () => {
  hasChanges.value = count.value !== props.data.count
})

function saveChange() {
  emits('changeCount', count.value)
  hasChanges.value = false
}

function decrementCount() {
  if (count.value > 1) {
    count.value -= 1
  }
}

function incrementCount() {
  count.value += 1
}
</script>

<template>
  <div class="order-item">
    <button class="remove-btn c-flex-all-center" @click="emits('destroy')">
      <BaseIcon :icon="icons.pages.basket.close" />
    </button>
    <div class="order-info">
      <div class="img">
        <ImageFrameCard
          v-if="typeof props.data.img === 'string' && props.data.img !== ''"
          :img="props.data.img"
          :alt="props.data.alt"
        />
        <div v-else class="img-placeholder c-flex-all-center">
          <span>{{ props.data.name.slice(0, 1).toUpperCase() }}</span>
        </div>
      </div>
      <div class="item c-flex-all-center">
        <div class="sub-item sub-1">
          <div class="name">
            <h2>{{ props.data.name }}</h2>
          </div>
          <div class="info">
            <h3>price: {{ props.data.price }}</h3>
            <h3>total: {{ props.data.price * count }}</h3>
          </div>
        </div>
        <div class="sub-item sub-2">
          <div class="change-box c-flex-all-center">
            <button
              class="change-btn change-box-item"
              @click="decrementCount"
              :disabled="count <= 1"
            >
              <BaseIcon
                :icon="icons.pages.basket.minus"
                stroke-color="--c-v-7"
                fill-color="--c-v-7"
              />
            </button>
            <span class="change-box-item">{{ count }}</span>
            <button class="change-btn change-box-item" @click="incrementCount">
              <BaseIcon
                :icon="icons.pages.basket.plus"
                stroke-color="--c-v-7"
                fill-color="--c-v-7"
              />
            </button>
          </div>
          <button
            v-if="hasChanges"
            class="save-btn not-saved"
            @click="saveChange"
          >
            <span>Save changes</span>
          </button>
          <span v-else class="save-btn saved c-flex-all-center">Saved</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.order-item {
  @apply w-full max-w-200 min-h-30 flex flex-row justify-between
    hover:backdrop-brightness-90 pt-4;
}

.order-item > .remove-btn {
  @apply cursor-pointer w-7/100 h-full p-2
    hover:bg-(--c-v-1)/10;
}

.order-item > .order-info {
  @apply flex flex-row items-center justify-between
    w-91/100 h-full;
}

.order-item > .order-info > .img {
  @apply size-19 shrink-0;
}

.order-item > .order-info > .img .img-placeholder {
  @apply size-full rounded-xl font-bold text-2xl
    bg-(--c-v-6) text-(--c-v-4);
}

.order-item > .order-info > .item {
  @apply flex flex-row gap-2 size-full;
}

.order-info .sub-item {
  @apply flex flex-col justify-center
    h-full gap-2;
}

.order-info .sub-1 {
  @apply px-2 flex-1 min-w-0;
}

.order-info .sub-2 {
  @apply items-center w-40/100 flex-col;
}

.sub-item .name {
  @apply h-fit w-fit;
}

.order-info h2 {
  @apply text-lg md:text-xl text-nowrap text-ellipsis overflow-hidden;
}

.order-info .change-box {
  @apply flex flex-row w-fit h-fit
    rounded-xl border overflow-hidden
    border-(--c-v-1);
}

.order-info .change-box-item {
  @apply flex justify-center items-center
    size-10;
}

.order-item .change-btn {
  @apply cursor-pointer
    bg-(--c-v-1);
}

.order-item .change-btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.sub-item .info {
  @apply flex flex-row gap-2 md:gap-4;
}

.sub-item h3 {
  @apply w-fit h-fit text-nowrap text-sm md:text-base;
}

.sub-item .save-btn {
  @apply w-full md:w-65/100 h-10 border rounded-xl font-bold
    bg-none;
}

.sub-item .not-saved {
  @apply bg-(--c-v-1)
    text-(--c-v-7)
    cursor-pointer;
}

.sub-item .saved {
  @apply text-(--c-v-4) cursor-default;
}

@media (max-width: 640px) {
  .order-item {
    @apply flex-col min-h-auto;
  }

  .order-item > .remove-btn {
    @apply w-full p-4;
  }

  .order-item > .order-info {
    @apply w-full flex-col h-auto;
  }

  .order-item > .order-info > .item {
    @apply flex-col w-full gap-4 p-4;
  }

  .order-info .sub-1 {
    @apply w-full p-4;
  }

  .order-info .sub-2 {
    @apply w-full flex flex-col justify-between;
  }
}
</style>
