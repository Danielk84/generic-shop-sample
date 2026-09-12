<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'

const props = defineProps<{
  id: string
}>()

const store = useStore()
const router = useRouter()

const { mutate, isPending } = useMutation({
  mutationFn: async () => {
    return api.delete(`products/${props.id}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess: () => {},
  onError: (error) => {
    errorStatusHandler(error, router)
  },
})
</script>

<template>
  <div>
    <button :class="{ 'is-pending': isPending }" @click="mutate()">
      Delete
    </button>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
