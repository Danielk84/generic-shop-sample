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

watch(page, async (v: number) => {
  window.scroll(0, 0)
  router.push({
    name: props.pageName,
    params: route.params,
    query: { ...route.query, page: v },
  })
})
</script>

<template>
  <div class="pagination" v-if="props.last > 1">
    <button
      class="main-btn btn rounded-l-xl border-y border-(--c-v-10)"
      type="button"
      aria-label="Previous page"
      v-bind:class="{ off: page === 1 }"
      @click="setPage(page - 1)"
    >
      <BaseIcon
        :icon="icons.ui.pagination.previous"
        stroke-color="--c-v-7"
        fill-color="--c-v-7"
      />
    </button>
    <div v-for="i of pageRange(page, props.last)" :key="i">
      <button
        class="btn item"
        type="button"
        :aria-label="`Go to page ${i}`"
        :aria-current="i === page ? 'page' : undefined"
        v-bind:class="{ select: i === page }"
        @click="setPage(i)"
      >
        {{ i }}
      </button>
    </div>
    <button
      class="main-btn btn rounded-r-xl border-y border-(--c-v-10)"
      type="button"
      aria-label="Next page"
      :class="{ off: page === props.last || props.last === 0 }"
      @click="setPage(page + 1)"
    >
      <BaseIcon
        :icon="icons.ui.pagination.next"
        stroke-color="--c-v-7"
        fill-color="--c-v-7"
      />
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.pagination {
  @apply flex flex-row justify-center items-center
    size-fit rounded-xl overflow-hidden;
}

.pagination .main-btn {
  @apply bg-(--c-v-1) cursor-pointer transition-colors duration-200
    focus-visible:outline-2 focus-visible:outline-offset-2;
}

.pagination .btn {
  @apply min-w-11 min-h-11 px-4 cursor-pointer
    transition-colors duration-200
    focus-visible:outline-2 focus-visible:outline-offset-2
    text-(--c-v-1);
}

.pagination .item {
  @apply border-l border-y border-(--c-v-1);
}

.pagination .select {
  @apply bg-(--c-v-1) text-(--c-v-7);
}

.pagination .off {
  @apply brightness-75 pointer-events-none;
}
</style>
