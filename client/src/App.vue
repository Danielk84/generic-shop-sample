<script setup lang="ts">
import { computed, type Component } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import DefaultLayout from '@/layout/DefaultLayout.vue'
import { setLayout, type LayoutName } from '@/layout/layout'

const route = useRoute()

const layoutList = computed<LayoutName[]>(() => {
  const result: LayoutName[] = []

  // max layout components
  for (let count = 0; count < 3; count++) {
    const layout = route.meta[`layout_${count}`]
    if (layout === undefined) {
      break
    }
    result.push(layout as LayoutName)
  }

  return result
})

const comp = computed<Component>(() => {
  return setLayout(layoutList.value, DefaultLayout)
})
</script>

<template>
  <component :is="comp">
    <RouterView />
  </component>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
