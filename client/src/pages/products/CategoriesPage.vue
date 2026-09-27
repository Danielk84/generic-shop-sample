<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'

import type { Category } from '@/contracts/categories/response.interface'

const CategoriesView = defineAsyncComponent(
  () => import('@/components/common/categories/CategoriesView.vue'),
)

const router = useRouter()
</script>

<template>
  <CategoriesView
    @tag-handler="
      (category: Category) => {
        const q = category.tag.trim()
        router.push({ name: 'products-list', query: { q } })
      }
    "
  />
</template>
