<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'

const SearchBox = defineAsyncComponent(() => import('@/components/ui/search/SearchBox.vue'))
const ListPagination = defineAsyncComponent(() => import('@/components/ui/ListPagination.vue'))
const page = ref<number>(1)

const items = ref([])
</script>

<template>
  <div class="products-list">
    <div class="products-search">
      <SearchBox />
    </div>
    <div v-if="items.length === 0" class="empty">
      <h2>
        <span>Sory!</span>
        <span>But it is empty!</span>
      </h2>
    </div>
    <div v-else class="items">
    </div>
    <ListPagination
      :last=0
      @change-page="(newValue: number) => { page = newValue }"
    />
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.products-list {
  @apply w-full min-h-200 p-8
    flex flex-col items-center justify-between;
}

.products-list > .products-search {
  @apply w-80/100 h-fit;
}

.products-list > .empty > h2 {
  @apply text-4xl flex flex-col items-start justify-center;
}
</style>