<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
import type { Category } from '@/contracts/categories/response.interface'

const emits = defineEmits<{
  (e: 'tagHandler', category: Category): void
}>()

const router = useRouter()

const tagHandler = (category: Category) => {
  emits('tagHandler', category)
}

const TagCard = defineAsyncComponent(
  () => import('@/components/card/TagCard.vue'),
)

const isNotFound = ref<boolean>(false)
const { data, error, refetch } = useQuery({
  queryKey: ['categories-list'],
  queryFn: async () => api.get<Category[]>('categories/'),
  select: (res) => res.data,
  retry: 0,
})

defineExpose({
  refetch,
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          console.log('hello')
          isNotFound.value = true
        },
      })
    }
  },
  {
    immediate: true,
  },
)
</script>

<template>
  <div class="categories-view c-flex-all-center">
    <div v-if="isNotFound" class="not-found">There are not any categories.</div>
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
