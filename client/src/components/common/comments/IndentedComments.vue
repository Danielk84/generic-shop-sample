<script setup lang="ts">
import { ref } from 'vue'

import type { RelatedCommentsRequest } from '@/contracts/comments/request.schema'

const props = defineProps<{
  relation: RelatedCommentsRequest
  childrenAmount: number
}>()

const isOpen = ref<boolean>(false)
</script>

<template>
  <div v-if="props.childrenAmount > 0">
    <button @click="isOpen = !isOpen">
      <span>Show responses &lt;{{ props.childrenAmount }}&gt;</span>
    </button>
    <div v-if="isOpen">
      <!-- global components, see 'main.ts' -->
      <CommentsList
        :isParent="false"
        :relation="{
          parent: props.relation.parent,
          referrer: props.relation.referrer,
        }"
      />
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
