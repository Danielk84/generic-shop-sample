<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
// import { PC, type PCInput } from '@/contracts/categories/request.schema'
import { errorStatusHandler } from '@/utils/helper'

const props = defineProps<{
  id: string
}>()

const store = useStore()
const router = useRouter()

useMutation({
  mutationKey: ['admin-products-set-tags', props.id],
  mutationFn: async () => {
    return api.post(`categories/pc/${props.id}`, null, {
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
  <div></div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
