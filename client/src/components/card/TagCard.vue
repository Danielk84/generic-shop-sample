<script setup lang="ts">
import type { StyleValue } from 'vue'

import { getCssVar } from '@/utils/helper'
import type { Category } from '@/contracts/categories/response.interface'

const colorList: StyleValue = [
  {
    backgroundColor: getCssVar('--color-tag-card-bg-1'),
  },
  {
    backgroundColor: getCssVar('--color-tag-card-bg-2'),
  },
  {
    backgroundColor: getCssVar('--color-tag-card-bg-3'),
  },
  {
    backgroundColor: getCssVar('--color-tag-card-bg-4'),
  },
  {
    backgroundColor: getCssVar('--color-tag-card-bg-5'),
  },
] as const

const props = defineProps<{
  id: number
  tag: string
  palette: number
}>()

const emits = defineEmits<{
  (e: 'tagHandler', category: Category): void
}>()

if (props.palette < 0) {
  throw new Error('Palette must be bigger than zero.')
}
</script>

<template>
  <button
    @click="emits('tagHandler', { id: props.id, tag: props.tag })"
    class="tag"
    :style="colorList[props.palette % colorList.length]"
  >
    {{ props.tag }}
  </button>
</template>

<style scoped>
@reference "@/styles/index.css";

.tag {
  @apply w-62.5 h-20 overflow-hidden
    text-xl font-bold rounded-2xl p-4
    cursor-pointer
    text-tag-card-text text-nowrap text-ellipsis;
}
</style>
