<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

import type { Category } from '@/contracts/categories/response.interface'

const TagCard = defineAsyncComponent(
  () => import('@/components/card/TagCard.vue'),
)

const props = defineProps<{ data?: Category[] }>()
const emits = defineEmits<{
  (e: 'tagHandler', category: Category): void
}>()

const tagHandler = (category: Category) => {
  emits('tagHandler', category)
}
</script>

<template>
  <div class="categories-view c-flex-all-center">
    <div
      v-if="props.data === undefined || props.data.length === 0"
      class="not-found"
    >
      There are not any categories.
    </div>
    <div v-else class="tags">
      <TagCard
        v-for="tag of data"
        :key="tag.id"
        :id="tag.id"
        :tag="tag.tag"
        :palette="tag.id"
        @tag-handler="tagHandler"
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.categories-view {
  @apply w-full h-full p-10;
}

.categories-view .not-found {
  @apply text-2xl font-bold;
}

.tags {
  @apply flex flex-row flex-wrap gap-4
    place-content-center;
}
</style>
