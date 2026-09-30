<script setup lang="ts">
import type { StyleValue } from 'vue'

import { getCssVar } from '@/utils/helper'
import type { TagCardProps } from '@/components/card/types'
import type { Category } from '@/contracts/categories/response.interface'

const colorList: StyleValue = [
  {
    backgroundColor: getCssVar('--c-v-11'),
  },
  {
    backgroundColor: getCssVar('--c-v-12'),
  },
  {
    backgroundColor: getCssVar('--c-v-13'),
  },
  {
    backgroundColor: getCssVar('--c-v-14'),
  },
  {
    backgroundColor: getCssVar('--c-v-15'),
  },
] as const

const props = defineProps<TagCardProps>()

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
    cursor-pointer hover:brightness-110
    text-(--c-v-1) text-nowrap text-ellipsis;
}
</style>
