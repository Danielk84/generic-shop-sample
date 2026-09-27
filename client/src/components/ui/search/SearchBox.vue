<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import icons from '@/utils/icons'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const route = useRoute()
const router = useRouter()
const query = ref<string>('')

watch(
  () => route.query.q,
  (q) => {
    query.value = typeof q === 'string' ? q : ''
  },
  { immediate: true },
)

function onSearch(event: Event) {
  event.preventDefault()
  const q = query.value.trim()
  const target =
    route.name === 'products-popular' ? 'products-popular' : 'products-list'
  if (q === '') {
    router.push({ name: target })
    return
  }
  router.push({ name: target, query: { q } })
}
</script>

<template>
  <form
    class="search-box c-flex-all-center"
    role="search"
    @submit.prevent="onSearch"
  >
    <label for="search" class="sr-only">Search products</label>
    <input
      v-model="query"
      class="search-input"
      id="search"
      type="search"
      placeholder="Search"
      autocomplete="off"
    />
    <button class="search-btn" type="submit" aria-label="Search">
      <BaseIcon
        :icon="icons.ui.search.btn"
        fill-color="none"
        fill-dark-color="none"
        stroke-color="--c-v-7"
      />
    </button>
  </form>
</template>

<style scoped>
@reference "@/styles/index.css";

.search-box {
  @apply w-full h-14 flex-row
    rounded-4xl p-1.5 bg-(--c-v-1)
    border-2 border-(--c-v-6);
}

.search-box > .search-input {
  @apply w-full h-full focus:outline-none
    mx-3 text-(--c-v-6) text-xl min-h-11;
}

.search-box > input::placeholder {
  color: var(--c-v-6);
}

.search-box > .search-btn {
  @apply w-11 h-11 shrink-0 flex items-center justify-center rounded-full
    bg-(--c-v-4) cursor-pointer transition-colors duration-200
    focus-visible:outline-2 focus-visible:outline-offset-2;
}

.sr-only {
  @apply absolute w-px h-px p-0 -m-px overflow-hidden whitespace-nowrap border-0;
}
</style>
