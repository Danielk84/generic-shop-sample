<script setup lang="ts">
import { defineAsyncComponent, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import icons from '@/utils/icons'
import { range } from '@/utils/helper'
const BaseIcon = defineAsyncComponent(() => import('@/components/ui/BaseIcon.vue'))

const props = defineProps<{ last: number }>()
const emits = defineEmits<{
  (e: 'changePage', page: number): void,
}>()
const route = useRoute()

const page = ref<number>(0)

const pageRange = (page: number, last: number) => {
  const windowSize = 3
  let start = Math.max(1, page - 1)
  let end = Math.min(last, start + windowSize - 1)
  
  if (end - start < windowSize - 1) {
    start = Math.max(1, end - windowSize + 1)
  }
  
  return range(start, end + 1)
}


const setPage = (value: number) => {
  if (value > props.last || value < 1) {
    return
  }
  page.value = value
}

onMounted(async () => {
  page.value = Number(route.query.page)
  if (isNaN(page.value) || page.value > props.last || page.value < 1) {
    page.value = 1
  }
  emits('changePage', page.value)
})

watch(page, async (newValue: number) => {
  emits('changePage', newValue)
})
</script>

<template>
  <div class="pagination">
    <button
      class="main-btn btn"
      v-bind:class="{ 'off': page === 1 }"
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
        v-bind:class="{ 'select': i === page }"
        @click="setPage(i)"
      >
        {{ i }}
      </button>
    </div>
    <button
      class="main-btn btn item"
      v-bind:class="{ 'off': page === props.last || props.last === 0 }"
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