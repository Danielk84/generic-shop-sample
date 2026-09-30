<script setup lang="ts">
import { defineAsyncComponent, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { errorStatusHandler } from '@/utils/helper'
import type { Category } from '@/contracts/categories/response.interface'

const TagsListView = defineAsyncComponent(
  () => import('@/components/common/categories/TagsListView.vue'),
)

const emits = defineEmits<{
  (e: 'tagHandler', category: Category): void
}>()

const router = useRouter()

const { data, error, refetch } = useQuery({
  queryKey: ['categories-list'],
  queryFn: async () => api.get<Category[]>('categories/'),
  select: (res) => res.data,
  retry: 0,
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {
          // empty block
          return
        },
      })
    }
  },
  { immediate: true },
)

defineExpose({
  refetch,
})

const tagHandler = (category: Category) => {
  emits('tagHandler', category)
}
window.scroll(0, 0)
</script>

<template>
  <TagsListView :data="data" @tag-handler="tagHandler" />
</template>
