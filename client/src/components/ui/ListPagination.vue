<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import icons from '@/utils/icons'
import { range } from '@/utils/helper'
const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const props = defineProps<{
  last: number
  pageName: string
}>()
const emits = defineEmits<{
  (e: 'changePage', page: number): void
}>()

const route = useRoute()
const router = useRouter()

const page = ref<number>(1)

function pageRange(page: number, last: number) {
  const windowSize = 3
  let start = Math.max(1, page - 1)
  let end = Math.min(last, start + windowSize - 1)

  if (end - start < windowSize - 1) {
    start = Math.max(1, end - windowSize + 1)
  }

  return range(start, end + 1)
}

function setPage(value: number) {
  if (value > props.last || value < 1) {
    return
  }
  page.value = value
}

function syncPageFromRoute() {
  const p = Number(route.query.page)
  if (!Number.isInteger(p) || p < 1) {
    page.value = 1
    return
  }

  if (props.last > 1 && p > props.last) {
    page.value = props.last
    return
  }

  page.value = p
}

watch(
  () => [route.query.page, props.last],
  () => {
    syncPageFromRoute()
    emits('changePage', page.value)
  },
  { immediate: true },
)

watch(
  page,
  async (v: number) => {
    window.scroll(0, 0)
    router.push({ name: props.pageName, query: { page: v } })
  },
)
</script>

<template>
  <div class="pagination" v-if="props.last > 1">
    <button
      class="main-btn btn"
      v-bind:class="{ off: page === 1 }"
      @click="setPage(page - 1)"
    >
      <BaseIcon
        :icon="icons.ui.pagination.previous"
        stroke-color="--color-pagination-icon"
        fill-color="--color-pagination-icon"
      />
    </button>
    <div v-for="i of pageRange(page, props.last)" :key="i">
      <button
        class="btn item"
        v-bind:class="{ select: i === page }"
        @click="setPage(i)"
      >
        {{ i }}
      </button>
    </div>
    <button
      class="main-btn btn item"
      :class="{ off: page === props.last || props.last === 0 }"
      @click="setPage(page + 1)"
    >
      <BaseIcon
        :icon="icons.ui.pagination.next"
        stroke-color="--color-pagination-icon"
        fill-color="--color-pagination-icon"
      />
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.pagination {
  @apply flex flex-row justify-center items-center
    size-fit rounded-xl overflow-hidden
    border border-pagination-border;
}

.pagination .main-btn {
  @apply bg-pagination-btn;
}

.pagination .btn {
  @apply hover:cursor-pointer min-w-10 min-h-10 px-4
    text-pagination-default-text;
}

.pagination .item {
  @apply border-l border-pagination-border;
}

.pagination .select {
  @apply bg-pagination-select-bg text-pagination-select-text;
}

.pagination .off {
  @apply brightness-75 pointer-events-none;
}
</style>
